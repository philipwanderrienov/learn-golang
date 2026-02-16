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
	// user repository and service
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc)

	// church member repository and service
	churchRepo := repository.NewChurchMemberRepository(db)
	churchSvc := service.NewChurchMemberService(churchRepo)
	churchHandler := handler.NewChurchMemberHandler(churchSvc)

	r := mux.NewRouter()

	// User routes
	r.HandleFunc("/users", userHandler.CreateUserHandler).Methods("POST")
	r.HandleFunc("/users", userHandler.ListUsersHandler).Methods("GET")
	r.HandleFunc("/users/{id}", userHandler.GetUserHandler).Methods("GET")
	r.HandleFunc("/users/{id}", userHandler.UpdateUserHandler).Methods("PUT")
	r.HandleFunc("/users/{id}", userHandler.DeleteUserHandler).Methods("DELETE")

	// Church member routes
	r.HandleFunc("/members", churchHandler.CreateMemberHandler).Methods("POST")
	r.HandleFunc("/members", churchHandler.ListMembersHandler).Methods("GET")
	r.HandleFunc("/members/joined", churchHandler.ListMembersByDateHandler).Methods("GET")
	r.HandleFunc("/members/{id}", churchHandler.GetMemberHandler).Methods("GET")
	r.HandleFunc("/members/{id}", churchHandler.UpdateMemberHandler).Methods("PUT")
	r.HandleFunc("/members/{id}", churchHandler.DeleteMemberHandler).Methods("DELETE")

	// swagger UI
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Apply middleware (similar to .NET's middleware pipeline)
	handler := middleware.RecoveryMiddleware(middleware.LoggingMiddleware(r))

	log.Printf("starting server on %s", addr)
	return http.ListenAndServe(addr, handler)
}
