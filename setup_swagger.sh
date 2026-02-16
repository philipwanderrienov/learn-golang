#!/bin/bash
# Quick Swagger setup script
# Run this once to set up Swagger UI

set -e

echo "Installing swag CLI..."
go install github.com/swaggo/swag/cmd/swag@latest

echo "Adding swagger dependencies..."
go get -u github.com/swaggo/http-swagger
go get -u github.com/swaggo/swag
go mod tidy

echo "Generating swagger docs..."
swag init -g cmd/users/main.go -o docs

echo "✓ Swagger setup complete!"
echo ""
echo "Run the app:"
echo "  go run ./cmd/users"
echo ""
echo "Then open:"
echo "  http://localhost:8080/swagger/index.html"
