package config

import (
	"os"
)

type Config struct {
	HTTPAddr string
}

func Load() *Config {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "3010"
	}

	return &Config{
		HTTPAddr: ":" + port,
	}
}


