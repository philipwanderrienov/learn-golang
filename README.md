# Users microservice (Go) — simple CRUD

This is a tiny Users microservice implemented in Go with a layered (controller/service/repository) structure inspired by common .NET patterns.

Quick notes
- Service: single microservice handling `User` CRUD.
- DB: PostgreSQL via `database/sql` + `github.com/lib/pq`.
- Run: set `DB_CONN` then `go run ./cmd/users`.

Setup
1. Create a Postgres database locally and run the migration in `migrations/001_create_users.sql`.
2. Copy `.env.example` to `.env` or export `DB_CONN` in your shell.

Run
```bash
export DB_CONN="postgres://user:pass@localhost:5432/mydb?sslmode=disable"
go run ./cmd/users
```

API
- POST /users — body: {"name":"...","email":"..."} -> returns {"id":123}
- GET /users/{id} — returns user JSON or 404
- PUT /users/{id} — body: {"name":"...","email":"..."} -> 204
- DELETE /users/{id} — 204
- GET /users — returns array of users

Testing the API

**Option 1: Swagger UI (interactive, recommended)**
See [SWAGGER_SETUP.md](SWAGGER_SETUP.md) for one-time setup.
Once setup, open http://localhost:8080/swagger/index.html

**Option 2: cURL**
```bash
curl -X POST -H "Content-Type: application/json" -d '{"name":"Alice","email":"alice@example.com"}' http://localhost:8080/users
curl http://localhost:8080/users
```

If you'd like, next steps I can help with:
- Add Dockerfile and docker-compose for Postgres + service
- Add migrations runner (golang-migrate) and integration tests
- Add logging/middleware or JWT auth
