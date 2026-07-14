package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/sushilchlgn/go-url-shortener/internal/store"
)

// codeFromPath extracts the short code from the request path. Netlify's
// redirect rewrites "/r/<code>" to "/.netlify/functions/redirect/<code>",
// so the code is always the final path segment.
func codeFromPath(path string) string {
	trimmed := strings.TrimRight(path, "/")
	parts := strings.Split(trimmed, "/")
	return parts[len(parts)-1]
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	code := codeFromPath(req.Path)
	if code == "" || code == "redirect" {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "missing short code",
		}, nil
	}

	db, err := store.Open(ctx)
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "database unavailable",
		}, nil
	}
	defer db.Close()

	target, err := store.Resolve(ctx, db, code)
	if errors.Is(err, sql.ErrNoRows) {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       "short link not found",
		}, nil
	}
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "lookup failed",
		}, nil
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusFound,
		Headers:    map[string]string{"Location": target},
	}, nil
}

func main() {
	lambda.Start(handler)
}