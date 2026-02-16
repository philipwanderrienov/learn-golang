# Debug Guide: Application Flow & API Request Tracing

This guide shows you the exact flow from app startup through an API request, explaining each step.

---

## 1. Application Startup Flow

### Step 1a: `go run ./cmd/users` (Entry Point)
**File:** [cmd/users/main.go](cmd/users/main.go)

```go
func main() {
    // Step 1: Load configuration
    configPath := "config/appsettings.json"
    var conf *cfg.Config
    if _, err := os.Stat(configPath); err == nil {
        c, err := cfg.Load(configPath)
        // cfg.Load() reads JSON and allows env overrides
        // See: pkg/db/config/config.go
        if err != nil {
            log.Fatalf("failed to load config: %v", err)
        }
        conf = c
    } else {
        // fallback to env vars if no JSON file
        conf = &cfg.Config{}
        conf.Database.ConnectionString = os.Getenv("DB_CONN")
        conf.Server.Addr = os.Getenv("ADDR")
    }
```

**What happens:**
- Reads `config/appsettings.json` (or falls back to `DB_CONN` env var).
- Parses JSON into a `Config` struct with `Database.ConnectionString` and `Server.Addr`.
- Env vars (`DB_CONN`, `ADDR`) override JSON if set.

**Debug tip:** Add a log after config load to see what was loaded:
```go
log.Printf("Loaded config: DB=%s, Addr=%s", conf.Database.ConnectionString, conf.Server.Addr)
```

---

### Step 1b: Connect to PostgreSQL Database
**File:** [pkg/db/db.go](pkg/db/db.go)

```go
// In main.go:
if conf.Database.ConnectionString == "" {
    log.Fatal("database connection string is required...")
}

dbConn, err := db.ConnectDB(conf.Database.ConnectionString)
if err != nil {
    log.Fatalf("failed to connect to db: %v", err)
}
defer dbConn.Close()  // Close connection when app exits
```

**db.ConnectDB() does:**
```go
func ConnectDB(connStr string) (*sql.DB, error) {
    // sql.Open loads the pq driver (postgres)
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }

    // Ping tests connectivity early
    if err := db.Ping(); err != nil {
        db.Close()
        return nil, err
    }
    return db, nil  // Return *sql.DB pool
}
```

**What happens:**
- `sql.Open()` creates a connection pool (doesn't connect yet).
- `db.Ping()` actually connects and validates credentials.
- If credentials/DB don't exist → error (e.g., "role public does not exist").

**Debug tip:** If you see "connection refused" → Postgres isn't running.
If you see "role X does not exist" → wrong username.

---

### Step 1c: Wire Dependency Injection (repo → service → handlers)
**File:** [internal/server/server.go](internal/server/server.go)

```go
// In main.go:
addr := conf.Server.Addr
if addr == "" {
    addr = ":8080"
}

if err := server.Run(addr, dbConn); err != nil {
    log.Fatalf("server stopped with error: %v", err)
}
```

**server.Run() does:**
```go
func Run(addr string, db *sql.DB) error {
    // 1. Create repository (low-level DB queries)
    repo := repository.NewUserRepository(db)
    // repo is a struct that holds *sql.DB, calls db.QueryRowContext, etc.

    // 2. Create service (business logic)
    svc := service.NewUserService(repo)
    // svc is a struct that holds *repository.UserRepository
    // Delegates all operations to repo

    // 3. Create HTTP handlers (controllers)
    h := handler.NewUserHandler(svc)
    // h is a struct that holds *service.UserService
    // Calls svc methods like CreateUser(), GetUser(), etc.

    // 4. Wire HTTP routes
    r := mux.NewRouter()
    r.HandleFunc("/users", h.CreateUserHandler).Methods("POST")
    r.HandleFunc("/users", h.ListUsersHandler).Methods("GET")
    r.HandleFunc("/users/{id}", h.GetUserHandler).Methods("GET")
    r.HandleFunc("/users/{id}", h.UpdateUserHandler).Methods("PUT")
    r.HandleFunc("/users/{id}", h.DeleteUserHandler).Methods("DELETE")

    // 5. Start HTTP server
    log.Printf("starting server on %s", addr)
    return http.ListenAndServe(addr, r)
    // Server now listens on :8080, waiting for HTTP requests
}
```

**Dependency flow:**
```
PostgreSQL DB
    ↓
*sql.DB (connection pool)
    ↓
UserRepository (queries)
    ↓
UserService (business logic)
    ↓
UserHandler (HTTP handlers)
    ↓
Gorilla Router
    ↓
HTTP Server listening on :8080
```

**Debug tip:** Server is now listening. You should see `starting server on :8080` in logs.

---

## 2. API Request Flow (Example: POST /users)

### Step 2a: Client sends HTTP request
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}' http://localhost:8080/users
```

---

### Step 2b: Router matches route → calls CreateUserHandler
**File:** [internal/handler/user_handler.go](internal/handler/user_handler.go)

```go
func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    // Step 2b-1: Decode JSON body INTO a User struct
    var in model.User
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        // If JSON is invalid, return 400 Bad Request
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    // in now = User{Name: "Alice", Email: "alice@example.com"}

    // Step 2b-2: Call service to create user
    id, err := h.svc.CreateUser(r.Context(), &in)
    if err != nil {
        // If service returns error, return 500 Internal Server Error
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Step 2b-3: Write success response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)  // HTTP 201
    json.NewEncoder(w).Encode(map[string]int64{"id": id})
    // Response: {"id": 1}
}
```

**What happens:**
1. JSON body is decoded into `model.User` struct.
2. Handler calls service to create user (delegates business logic).
3. Service returns new user ID (or error).
4. Handler writes HTTP 201 + JSON response.

**Debug tip:** If you see 400 Bad Request → JSON format is wrong.
If you see 500 → database error (check server logs).

---

### Step 2c: Service processes create request
**File:** [internal/service/user_service.go](internal/service/user_service.go)

```go
func (s *UserService) CreateUser(ctx context.Context, u *model.User) (int64, error) {
    // In a .NET style you'd validate DTOs here; keep simple and delegate to repo.
    // For now, just pass through to repository (no validation)
    return s.repo.Create(ctx, u)
}
```

**What happens:**
- Service receives `context.Context` (used for timeouts/cancellation).
- Service delegates to repository's `Create()` method.
- In a real app, you'd add validation (e.g., email format check) here.

**Debug tip:** Add logging here to trace the flow:
```go
log.Printf("[Service] CreateUser called with name=%s email=%s", u.Name, u.Email)
```

---

### Step 2d: Repository executes SQL INSERT
**File:** [internal/repository/user_repository.go](internal/repository/user_repository.go)

```go
func (r *UserRepository) Create(ctx context.Context, u *model.User) (int64, error) {
    // Step 2d-1: Get current timestamp
    now := time.Now().UTC()

    // Step 2d-2: Prepare SQL query
    // This is a Postgres-specific parameterized query ($1, $2, $3)
    // Parameters prevent SQL injection
    var id int64
    err := r.db.QueryRowContext(ctx,
        `INSERT INTO users (name, email, created_at) VALUES ($1, $2, $3) RETURNING id`,
        u.Name,      // $1 = "Alice"
        u.Email,     // $2 = "alice@example.com"
        now,         // $3 = 2026-02-16 11:31:36 UTC
    ).Scan(&id)      // RETURNING id captures the auto-generated ID
    // id now = 1

    // Step 2d-3: Check for errors (e.g., duplicate email)
    if err != nil {
        return 0, err  // e.g., "pq: duplicate key value violates unique constraint \"users_email_key\""
    }

    // Step 2d-4: Return new ID
    return id, nil
}
```

**What happens:**
1. Repository builds a parameterized SQL INSERT.
2. Executes it with `QueryRowContext()` (context allows timeout).
3. `RETURNING id` retrieves the auto-generated ID from Postgres.
4. `.Scan(&id)` reads the result into the `id` variable.
5. If no error → return ID to service → back to handler → to client.

**Actual SQL sent to Postgres:**
```sql
INSERT INTO users (name, email, created_at) VALUES ('Alice', 'alice@example.com', '2026-02-16 11:31:36 UTC') RETURNING id;
```

**Debug tip:** If you see "duplicate key" error → email already exists (UNIQUE constraint).
If you see "column X does not exist" → migration (SQL script) wasn't run.

---

## 3. Full Request/Response Timeline

```
1. Client: curl -X POST ... http://localhost:8080/users
   ↓
2. Router: matches /users + POST → calls CreateUserHandler
   ↓
3. Handler: decodes JSON body
   ↓
4. Handler: calls svc.CreateUser(ctx, user)
   ↓
5. Service: delegates to repo.Create(ctx, user)
   ↓
6. Repository: builds SQL + executes db.QueryRowContext()
   ↓
7. Postgres: validates constraints (unique email) + inserts row + returns ID
   ↓
8. Repository: scans ID, returns to service
   ↓
9. Service: returns ID to handler
   ↓
10. Handler: writes HTTP 201 + JSON {"id": 1}
    ↓
11. Client: receives response
```

---

## 4. How to Add Debug Logging

### Option A: Add logs to each layer
**File:** [internal/handler/user_handler.go](internal/handler/user_handler.go)
```go
func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    var in model.User
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    log.Printf("[Handler] CreateUserHandler: name=%s email=%s", in.Name, in.Email)

    id, err := h.svc.CreateUser(r.Context(), &in)
    if err != nil {
        log.Printf("[Handler] CreateUser failed: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    log.Printf("[Handler] User created: id=%d", id)
    ...
}
```

**File:** [internal/service/user_service.go](internal/service/user_service.go)
```go
func (s *UserService) CreateUser(ctx context.Context, u *model.User) (int64, error) {
    log.Printf("[Service] CreateUser called: name=%s email=%s", u.Name, u.Email)
    id, err := s.repo.Create(ctx, u)
    if err != nil {
        log.Printf("[Service] repo.Create failed: %v", err)
        return 0, err
    }
    log.Printf("[Service] repo.Create returned id=%d", id)
    return id, nil
}
```

**File:** [internal/repository/user_repository.go](internal/repository/user_repository.go)
```go
func (r *UserRepository) Create(ctx context.Context, u *model.User) (int64, error) {
    now := time.Now().UTC()
    log.Printf("[Repository] INSERT user: name=%s email=%s", u.Name, u.Email)
    var id int64
    err := r.db.QueryRowContext(ctx,
        `INSERT INTO users (name, email, created_at) VALUES ($1, $2, $3) RETURNING id`,
        u.Name, u.Email, now,
    ).Scan(&id)
    if err != nil {
        log.Printf("[Repository] SQL INSERT failed: %v", err)
        return 0, err
    }
    log.Printf("[Repository] INSERT returned id=%d", id)
    return id, nil
}
```

**Terminal output when you POST:**
```
[Handler] CreateUserHandler: name=Alice email=alice@example.com
[Service] CreateUser called: name=Alice email=alice@example.com
[Repository] INSERT user: name=Alice email=alice@example.com
[Repository] INSERT returned id=1
[Service] repo.Create returned id=1
[Handler] User created: id=1
```

### Option B: Use structured logging (optional upgrade)
For a .NET-like experience, use `github.com/sirupsen/logrus` or `go.uber.org/zap`:
```bash
go get github.com/sirupsen/logrus
```

Then:
```go
import "github.com/sirupsen/logrus"

log := logrus.New()
log.WithFields(logrus.Fields{
    "name":  in.Name,
    "email": in.Email,
}).Info("Creating user")
```

---

## 5. Testing All Endpoints with Debug Output

Add the logs above, then test:

```bash
# Create
curl -X POST -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}' http://localhost:8080/users
# Watch server logs: [Handler] → [Service] → [Repository]

# Get all
curl http://localhost:8080/users
# Logs: [Handler] ListUsersHandler → [Service] ListUsers → [Repository] List

# Get by ID
curl http://localhost:8080/users/1
# Logs: [Handler] GetUserHandler id=1 → [Service] GetUser → [Repository] GetByID

# Update
curl -X PUT -H "Content-Type: application/json" \
  -d '{"name":"Alice B","email":"aliceb@example.com"}' http://localhost:8080/users/1
# Logs: [Handler] UpdateUserHandler id=1 → [Service] UpdateUser → [Repository] Update

# Delete
curl -X DELETE http://localhost:8080/users/1
# Logs: [Handler] DeleteUserHandler id=1 → [Service] DeleteUser → [Repository] Delete
```

Each request will trace through handler → service → repository → Postgres and back.

---

## Summary

| Layer | Responsibility | File |
|-------|---|---|
| **Handler (HTTP)** | Decode request, call service, encode response | `internal/handler/user_handler.go` |
| **Service (Business)** | Validate, orchestrate, delegate to repo | `internal/service/user_service.go` |
| **Repository (Data)** | Build SQL, execute queries, return results | `internal/repository/user_repository.go` |
| **Model (Domain)** | Struct definitions (User, etc.) | `internal/model/user.go` |
| **DB (Infrastructure)** | Connection pool, Postgres driver | `pkg/db/db.go` |
| **Config (Setup)** | Load JSON/env, construct server | `pkg/db/config/config.go` + `cmd/users/main.go` |

This is a .NET-inspired layered architecture: controllers → services → repositories → data models.

