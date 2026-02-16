# Go Microservice Architecture - Mirroring .NET PaymentHub Design

This document maps your Go Users microservice to the .NET PaymentHub-SSSS architecture, showing how each pattern has been replicated or adapted.

---

## 1. LAYERED ARCHITECTURE COMPARISON

### .NET Structure
```
PaymentHub.Microservice.SSSS
├── Controllers/              (HTTP Handlers)
├── Services/                 (Business Logic)
├── Repository/               (Data Access)
├── Entity/Models/            (Domain Models)
└── Common/Utilities/         (Helpers, Filters)
```

### Go Structure
```
golang-project
├── internal/
│   ├── handler/              (HTTP handlers - analogous to Controllers)
│   ├── service/              (Business logic - analogous to Services)
│   ├── repository/           (Data access layer)
│   ├── model/                (Domain entities)
│   ├── middleware/           (Filters/Attributes - analogous to LogStatusLocalAttribute)
│   └── server/               (Startup/Configuration)
├── pkg/
│   └── db/                   (Database connection, UnitOfWork)
└── config/
    └── appsettings.json      (Configuration - like .NET's appsettings.json)
```

---

## 2. PATTERN-BY-PATTERN MAPPING

### A. Generic Repository Pattern

#### .NET Implementation
```csharp
public interface IRepository<T> where T : class
{
    T Get(Expression<Func<T, bool>> expression);
    IEnumerable<T> GetAll();
    IEnumerable<T> GetAll(Expression<Func<T, bool>> expression);
    void Add(T entity);
    void Update(T entity);
    void Remove(T entity);
    Task<T> GetAsync(Expression<Func<T, bool>> expression);
    // ... more methods
}

public class Repository<T> : IRepository<T> where T : class
{
    private readonly DbContext _dbContext;
    private readonly DbSet<T> _entitySet;
    
    public void Add(T entity) => _dbContext.Add(entity);
    public void Update(T entity) => _dbContext.Update(entity);
    // ... implementation
}
```

#### Go Implementation (Your Project)
```go
// Repository interface
type Repository interface {
    ScanRow(ctx context.Context, query string, scanFn func(*sql.Row) error, args ...interface{}) error
    ScanRows(ctx context.Context, query string, scanFn func(*sql.Rows) error, args ...interface{}) error
    ExecUpdate(ctx context.Context, query string, args ...interface{}) error
}

// Base repository
type BaseRepository struct {
    db *sql.DB
}

// Specific repository
type UserRepository struct {
    base *BaseRepository
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) (int64, error) {
    return r.base.ScanRow(ctx, insertQuery, scanFn, u.Name, u.Email)
}
```

**Key Similarities:**
- ✅ Generic base with common CRUD methods
- ✅ Specific repositories inherit/embed base
- ✅ Type-safe, reusable
- ✅ DRY (Don't Repeat Yourself)

---

### B. Service Layer with Validation

#### .NET Implementation
```csharp
public interface IAidService
{
    Task<IEnumerable<AidEntity>> GetAll();
    Task<AidEntity> GetById(long id);
    Task Create(AidEntity entity);
    Task Update(AidEntity entity);
}

public class AidService : IAidService
{
    private readonly IBaseRepository<AidEntity> _repository;
    private readonly IUnitOfWork _uow;
    
    public async Task Create(AidEntity entity)
    {
        // Validation
        if (entity == null) throw new ArgumentNullException(nameof(entity));
        
        // Create
        await _repository.Create(entity);
        await _uow.CommitAsync();
    }
}
```

#### Go Implementation (Your Project)
```go
type UserService struct {
    repo *repository.UserRepository
}

func (s *UserService) CreateUser(ctx context.Context, u *model.User) (int64, error) {
    // Could add validation here
    if u == nil || u.Email == "" {
        return 0, errors.New("invalid user data")
    }
    
    return s.repo.Create(ctx, u)
}
```

**Pattern:** Both layers delegate to repositories, both support validation.

---

### C. Unit of Work Pattern (Transaction Management)

#### .NET Implementation
```csharp
public interface IUnitOfWork
{
    Task Commit();
    Task Rollback();
}

public class UnitOfWork : IUnitOfWork
{
    private readonly DbContext _dbContext;
    
    public async Task Commit() => await _dbContext.SaveChangesAsync();
    public async Task Rollback() => _dbContext.Dispose();
}
```

#### Go Implementation (Your Project - NEW!)
```go
type UnitOfWork interface {
    Begin(ctx context.Context) error
    Commit(ctx context.Context) error
    Rollback(ctx context.Context) error
    Tx() *sql.Tx
}

type UnitOfWorkImpl struct {
    db *sql.DB
    tx *sql.Tx
}

func (uow *UnitOfWorkImpl) Begin(ctx context.Context) error {
    tx, err := uow.db.BeginTx(ctx, nil)
    uow.tx = tx
    return err
}

func (uow *UnitOfWorkImpl) Commit(ctx context.Context) error {
    return uow.tx.Commit()
}
```

**File:** `pkg/db/uow.go`

**Usage Example:**
```go
uow := db.NewUnitOfWork(dbConn)
if err := uow.Begin(ctx); err != nil {
    return err
}

// Do work...
if err := someOperation(); err != nil {
    uow.Rollback(ctx)
    return err
}

return uow.Commit(ctx)
```

---

### D. Standardized Response Model

#### .NET Implementation
```csharp
public class ResponseModel
{
    public bool success { get; set; }
    public string code { get; set; }        // "00" = success, "01" = error
    public string message { get; set; }
    public object result { get; set; }
}

// Usage in controller
return new JsonResult(new ResponseModel 
{ 
    success = true, 
    code = "00", 
    result = data 
});
```

#### Go Implementation (Your Project - NEW!)
```go
type ResponseModel struct {
    Success bool        `json:"success"`
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Result  interface{} `json:"result,omitempty"`
}

func NewSuccessResponse(result interface{}) *ResponseModel {
    return &ResponseModel{
        Success: true,
        Code:    "00",
        Message: "Success",
        Result:  result,
    }
}

// Usage in handler
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(model.NewSuccessResponse(result))
```

**File:** `internal/model/response.go`

---

### E. Middleware & Filters

#### .NET Implementation
```csharp
// Custom attribute for logging
public class LogStatusLocalAttribute : ServiceFilterAttribute
{
    public override void OnActionExecuting(ActionExecutingContext context)
    {
        _logger.LogInformation($"{context.HttpContext.Request.Method} {context.HttpContext.Request.Path}");
    }
}

// Applied to controller
[ServiceFilter(typeof(LogStatusLocalAttribute))]
public class BaseServiceController<T> : ControllerBase { }

// Middleware pipeline
app.UseAuthentication();
app.UseAuthorization();
app.MapControllers();
```

#### Go Implementation (Your Project - NEW!)
```go
// Logging middleware
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)
        next.ServeHTTP(w, r)
    })
}

// Recovery middleware (panic handling)
func RecoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("[PANIC] %v", err)
                w.WriteHeader(http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// Applied in server
handler := middleware.RecoveryMiddleware(middleware.LoggingMiddleware(r))
http.ListenAndServe(addr, handler)
```

**File:** `internal/middleware/logging.go`

---

### F. Dependency Injection

#### .NET Implementation
```csharp
builder.Services.AddScoped<IUnitOfWork, UnitOfWork>();
builder.Services.AddTransient(typeof(IRepository<>), typeof(Repository<>));
builder.Services.AddTransient(typeof(IAidService), typeof(AidService));
builder.Services.AddScoped<LogStatusLocalAttribute>();
```

#### Go Implementation (Your Project)
```go
// In server.Run()
repo := repository.NewUserRepository(db)
svc := service.NewUserService(repo)
h := handler.NewUserHandler(svc)

// All dependencies explicitly wired
```

**Note:** Go uses explicit composition rather than reflection-based DI. Your approach is more transparent and doesn't require a DI container.

---

## 3. CONFIGURATION MANAGEMENT

### .NET Approach
```json
// appsettings.json
{
  "ConnectionStrings": {
    "PaymentHubDbDbContext": "mssql_connection_string"
  },
  "CustomConfig": {
    "DBEngine": "mssql"
  },
  "Jwt": {
    "Issuer": "PaymentHub",
    "Key": "secret_key"
  }
}
```

### Go Approach (Your Project)
```json
// config/appsettings.json
{
  "Database": {
    "ConnectionString": "postgres://user:pass@localhost:5432/learngolang?sslmode=disable"
  },
  "Server": {
    "Addr": ":8080"
  }
}
```

**Note:** Your Go project uses a simpler config structure. If needed, this can be extended with more sections (JWT, Database engine selection, etc.).

---

## 4. ERROR HANDLING STRATEGY

### .NET Multi-Layer Approach
```
Controller Layer
    ↓ (catches exception)
    ↓ returns ResponseModel { success: false, code: "01", message: error }
ServiceLayer
    ↓ (catches, logs, re-throws)
Repository Layer
    ↓ (catches, logs, re-throws)
Database
```

### Go Approach (Your Project & Middleware)
```
Handler Layer (middleware/recovery.go recovers from panics)
    ↓
Service Layer (validates input)
    ↓
Repository Layer (executes queries)
    ↓
Database
```

**Middleware Enhancement:** Your new `RecoveryMiddleware` catches panics and returns a standardized error response (similar to .NET's exception filters).

---

## 5. KEY DIFFERENCES (Go vs .NET)

| Aspect | .NET | Go |
|--------|------|-----|
| **DI Container** | Built-in (AddTransient/Scoped) | Explicit constructor injection |
| **Interfaces** | Applied everywhere with `I` prefix | Used strategically |
| **Async** | `async/await`, `Task<T>` | `context.Context` + goroutines |
| **Error Handling** | Try-catch, exception filters | Error returns, panic recovery |
| **Config** | IConfiguration with typed access | JSON unmarshaling |
| **Middleware** | Pipeline with context objects | Nested handler functions |
| **Types** | ORM (Entity Framework) | SQL driver (database/sql) |

---

## 6. EXTENSION OPPORTUNITIES

To further align with your .NET patterns:

### A. Add Validation Layer
```go
// internal/validator/user_validator.go
type UserValidator interface {
    ValidateCreate(user *model.User) error
    ValidateUpdate(user *model.User) error
}

type UserValidatorImpl struct {}

func (v *UserValidatorImpl) ValidateCreate(user *model.User) error {
    if user.Email == "" {
        return errors.New("email is required")
    }
    // More validations...
    return nil
}
```

### B. Add Query Builder Helper
```go
// internal/helper/filter_helper.go
// Similar to .NET's FilterHelper<T> for building LINQ expressions
type FilterExpression struct {
    Field    string
    Operator string  // "eq", "gt", "lt", "contains"
    Value    interface{}
}

func BuildWhereClause(filters []FilterExpression) string {
    // Builds SQL WHERE clause dynamically
}
```

### C. Add Pagination Helper
```go
// internal/model/pagination.go
type Pagination struct {
    PageSize int
    PageNumber int
    TotalCount int
}

type PaginatedResult[T any] struct {
    Data       []T
    Pagination Pagination
}
```

### D. Add Logging Service
```go
// pkg/logger/logger.go
// Structured logging similar to Serilog
type Logger interface {
    Info(msg string, fields map[string]interface{})
    Error(msg string, err error)
    Debug(msg string, fields map[string]interface{})
}
```

---

## 7. QUICK START FOR FAMILIAR PATTERNS

If you're coming from .NET, here's what you'll recognize:

| .NET Concept | Go Equivalent | File |
|--------------|---|---|
| `IRepository<T>` | `Repository` interface + `BaseRepository` | `internal/repository/base_repository.go` |
| `BaseService<T>` | `UserService` embed pattern | `internal/service/user_service.go` |
| `Controller` | HTTP handler functions | `internal/handler/user_handler.go` |
| `ResponseModel` | `ResponseModel` struct | `internal/model/response.go` |
| `IUnitOfWork` | `UnitOfWork` interface | `pkg/db/uow.go` |
| Custom Attributes | Middleware functions | `internal/middleware/logging.go` |
| `appsettings.json` | `config/appsettings.json` | `config/appsettings.json` |
| Dependency Injection | Explicit composition in `server.Run()` | `internal/server/server.go` |

---

## 8. NEXT STEPS

1. ✅ Generic Repository with interfaces
2. ✅ Service layer with business logic
3. ✅ Standardized response model
4. ✅ Unit of Work (transaction management)
5. ✅ Middleware for logging and recovery
6. 🔄 **Next:** Add validation layer
7. 🔄 **Then:** Add pagination helpers
8. 🔄 **Then:** Add structured logging (Serilog equivalent)
9. 🔄 **Then:** Add JWT authentication middleware

---

This Go project now mirrors your .NET PaymentHub microservice architecture! All core patterns are present and adapted for Go's idioms.
