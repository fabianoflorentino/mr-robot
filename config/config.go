package config

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	if err := godotenv.Load("config/.env"); err != nil {
		log.Fatalf("error: to load env config: %v", err)
	}

	return nil
}
