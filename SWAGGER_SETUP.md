# Swagger/OpenAPI Setup Guide

This project now includes Swagger UI for interactive API testing (similar to Postman in-browser).

## One-Time Setup

1. **Install `swag` CLI tool** (required to generate swagger docs):
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

2. **Generate Swagger docs** (run from project root):
```bash
swag init -g cmd/users/main.go -o docs
```

This creates a `docs/` directory with `swagger.json` and `swagger.yaml`.

3. **Add swagger dependencies** (run from project root):
```bash
go get -u github.com/swaggo/http-swagger
go get -u github.com/swaggo/swag
go mod tidy
```

## Running the App

After setup, run as usual:

```bash
go run ./cmd/users
```

Then open in browser:

```
http://localhost:8080/swagger/index.html
```

## Testing via Swagger UI

1. Open http://localhost:8080/swagger/index.html
2. You'll see all endpoints: POST/GET/PUT/DELETE /users
3. Click on any endpoint to expand it
4. Click "Try it out"
5. Fill in parameters/body
6. Click "Execute"

## Example API Flow in Swagger UI

### Create User (POST /users):
- Click on "POST /users"
- Click "Try it out"
- In the request body, enter:
```json
{
  "name": "Alice",
  "email": "alice@example.com"
}
```
- Click "Execute"
- Response: `{"id": 1}`

### Get All Users (GET /users):
- Click on "GET /users"
- Click "Try it out"
- Click "Execute"
- Response: array of all users

### Get User by ID (GET /users/{id}):
- Click on "GET /users/{id}"
- Click "Try it out"
- Enter ID in the path parameter (e.g., 1)
- Click "Execute"

### Update User (PUT /users/{id}):
- Click on "PUT /users/{id}"
- Click "Try it out"
- Enter ID in path
- Enter updated data in body:
```json
{
  "name": "Alice B",
  "email": "aliceb@example.com"
}
```
- Click "Execute"

### Delete User (DELETE /users/{id}):
- Click on "DELETE /users/{id}"
- Click "Try it out"
- Enter ID
- Click "Execute"

## Updating Swagger Docs

After you modify handlers or add new endpoints, regenerate docs:

```bash
swag init -g cmd/users/main.go -o docs
```

Then restart the app:

```bash
# Stop current app (Ctrl+C)
# Restart:
go run ./cmd/users
```

The Swagger UI will reflect the changes.

## Notes

- Swagger annotations are in `cmd/users/main.go` (global API info) and `internal/handler/user_handler.go` (endpoint details).
- The `@` comments are parsed by `swag` tool to generate OpenAPI spec.
- No need to write swagger by hand — annotations keep docs in sync with code.
- Swagger UI is served from `/swagger/` path on the API.

## Troubleshooting

**Q: Swagger UI shows "404" or blank?**
- Make sure you ran `swag init` to generate docs
- Make sure `docs/` folder exists in project root
- Restart the app

**Q: How do I add a new endpoint to Swagger?**
- Add a handler function in `internal/handler/`.
- Add Swagger annotations (the `// @` comments).
- Run `swag init -g cmd/users/main.go -o docs` to regenerate.
- Restart app.

**Q: Can I use curl instead of Swagger UI?**
- Yes, all endpoints still work with curl. Swagger is optional, just adds a UI.

