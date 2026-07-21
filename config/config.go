package config

import "os"

type Config struct {
	Port            string
	TokenServiceURL string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	tokenURL := os.Getenv("TOKEN_SERVICE_URL")
	if tokenURL == "" {
		tokenURL = "http://127.0.0.1:8081"
	}
	return &Config{
		Port:            port,
		TokenServiceURL: tokenURL,
	}
}
