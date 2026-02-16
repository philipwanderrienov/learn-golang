// @title Users Microservice API
// @version 1.0
// @description Simple CRUD API for managing users
// @termsOfService http://swagger.io/terms/
// @host localhost:8080
// @basePath /
// @schemes http
package main

import (
	"log"
	"os"

	_ "github.com/example/golang-project/docs"
	"github.com/example/golang-project/internal/server"
	"github.com/example/golang-project/pkg/db"
	cfg "github.com/example/golang-project/pkg/db/config"
)

// main is the service entrypoint. It loads configuration (from config/appsettings.json
// or environment variables), connects to the DB and starts the HTTP server.
func main() {
	configPath := "config/appsettings.json"

	var conf *cfg.Config
	if _, err := os.Stat(configPath); err == nil {
		c, err := cfg.Load(configPath)
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
		conf = c
	} else {
		// fallback to environment variables
		conf = &cfg.Config{}
		conf.Database.ConnectionString = os.Getenv("DB_CONN")
		conf.Server.Addr = os.Getenv("ADDR")
	}

	if conf.Database.ConnectionString == "" {
		log.Fatal("database connection string is required (set DB_CONN or config/appsettings.json)")
	}

	dbConn, err := db.ConnectDB(conf.Database.ConnectionString)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer dbConn.Close()

	addr := conf.Server.Addr
	if addr == "" {
		addr = ":8080"
	}

	if err := server.Run(addr, dbConn); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}
