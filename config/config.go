package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	if err := godotenv.Load("config/.env"); err != nil {
		log.Fatalf("error: to load env config: %v", err)
	}

	fmt.Println("Environment variables loaded successfully.")

	return nil
}
