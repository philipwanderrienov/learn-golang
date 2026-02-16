package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Database struct {
		ConnectionString string `json:"ConnectionString"`
	} `json:"Database"`
	Server struct {
		Addr string `json:"Addr"`
	} `json:"Server"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	// Allow env overrides (recommended for secrets)
	if v := os.Getenv("DB_CONN"); v != "" {
		c.Database.ConnectionString = v
	}
	if v := os.Getenv("ADDR"); v != "" {
		c.Server.Addr = v
	}
	return &c, nil
}
