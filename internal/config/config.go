package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	DatabaseURL        string
	Port               string
	JWTSecret          string
	GithubClientId     string
	GithubClientSecret string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		Port:               os.Getenv("PORT"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		GithubClientId:     os.Getenv("GITHUB_CLIENT_ID"),
		GithubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
	}
}
