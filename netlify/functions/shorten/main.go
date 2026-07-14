package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/sushilchlgn/go-url-shortener/internal/shortcode"
	"github.com/sushilchlgn/go-url-shortener/internal/store"
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
	Original string `json:"original_url"`
}

const codeLength = 6
const maxAttempts = 5

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != http.MethodPost {
		return jsonResponse(http.StatusMethodNotAllowed, map[string]string{"error": "use POST"})
	}

	var body shortenRequest
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		return jsonResponse(http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
	}

	target := strings.TrimSpace(body.URL)
	parsed, err := url.ParseRequestURI(target)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return jsonResponse(http.StatusBadRequest, map[string]string{"error": "url must be a valid http(s) URL"})
	}

	db, err := store.Open(ctx)
	if err != nil {
		return jsonResponse(http.StatusInternalServerError, map[string]string{"error": "database unavailable"})
	}
	defer db.Close()

	var code string
	for attempt := 0; attempt < maxAttempts; attempt++ {
		code, err = shortcode.Generate(codeLength)
		if err != nil {
			return jsonResponse(http.StatusInternalServerError, map[string]string{"error": "could not generate code"})
		}
		if err = store.Insert(ctx, db, code, target); err == nil {
			break
		}
	}
	if err != nil {
		return jsonResponse(http.StatusInternalServerError, map[string]string{"error": "could not save link"})
	}

	base := os.Getenv("PUBLIC_BASE_URL")
	if base == "" {
		base = "https://" + req.Headers["host"]
	}

	return jsonResponse(http.StatusCreated, shortenResponse{
		Code:     code,
		ShortURL: base + "/r/" + code,
		Original: target,
	})
}

func jsonResponse(status int, payload interface{}) (*events.APIGatewayProxyResponse, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(b),
	}, nil
}

func main() {
	lambda.Start(handler)
}
