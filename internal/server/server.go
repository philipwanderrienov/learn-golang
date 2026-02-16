package server

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/example/golang-project/internal/handler"
	"github.com/example/golang-project/internal/middleware"
	"github.com/example/golang-project/internal/repository"
	"github.com/example/golang-project/internal/service"
)

// Run wires dependencies (repo -> service -> handlers), sets up routes and starts the HTTP server.
func Run(addr string, db *sql.DB) error {
	// repository
	repo := repository.NewUserRepository(db)

	// service
	svc := service.NewUserService(repo)

	// handlers
	h := handler.NewUserHandler(svc)

	r := mux.NewRouter()

	// route definitions (REST-like)
	r.HandleFunc("/users", h.CreateUserHandler).Methods("POST")
	r.HandleFunc("/users", h.ListUsersHandler).Methods("GET")
	r.HandleFunc("/users/{id}", h.GetUserHandler).Methods("GET")
	r.HandleFunc("/users/{id}", h.UpdateUserHandler).Methods("PUT")
	r.HandleFunc("/users/{id}", h.DeleteUserHandler).Methods("DELETE")

	// swagger UI
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Apply middleware (similar to .NET's middleware pipeline)
	handler := middleware.RecoveryMiddleware(middleware.LoggingMiddleware(r))

	log.Printf("starting server on %s", addr)
	return http.ListenAndServe(addr, handler)
}
