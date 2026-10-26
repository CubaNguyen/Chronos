package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	ParserAddress string
	AIAddress     string
}

func Load() *Config {
	// Tự động load file .env nếu có
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port cho Go API
	}

	parserAddr := os.Getenv("PARSER_ADDRESS")
	if parserAddr == "" {
		parserAddr = "localhost:50051"
	}

	aiAddr := os.Getenv("AI_ADDRESS")
	if aiAddr == "" {
		aiAddr = "localhost:50052"
	}

	return &Config{
		Port:          port,
		ParserAddress: parserAddr,
		AIAddress:     aiAddr,
	}
}