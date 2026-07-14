package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/sushilchlgn/go-url-shortener/internal/store"
)

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	code := req.QueryStringParameters["code"]
	if code == "" {
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
