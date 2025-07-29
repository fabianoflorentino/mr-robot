package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	// Skip loading .env file if SKIP_ENV_FILE is set (useful for tests)
	if os.Getenv("SKIP_ENV_FILE") == "true" {
		fmt.Println("Skipping .env file loading (SKIP_ENV_FILE=true)")
		return nil
	}

	if err := godotenv.Load("config/.env"); err != nil {
		log.Fatalf("error: to load env config: %v", err)
	}

	fmt.Println("Environment variables loaded successfully.")

	return nil
}
