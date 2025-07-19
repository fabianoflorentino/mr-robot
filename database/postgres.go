package database

import (
	"errors"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	username string = os.Getenv("POSTGRES_USER")
	host     string = os.Getenv("POSTGRES_HOST")
	password string = os.Getenv("POSTGRES_PASSWORD")
	database string = os.Getenv("POSTGRES_DB")
	port     string = os.Getenv("POSTGRES_PORT")
	sslmode  string = os.Getenv("POSTGRES_SSLMODE")
	timezone string = os.Getenv("POSTGRES_TIMEZONE")
)

var (
	exists bool
	DB     *gorm.DB
)

func InitDB() error {
	dsn := setPostgresConnectionString()
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := pgCryptoExtension(db); err != nil {
		log.Fatalf("failed to enable pgcrypto extension: %v", err)
	}

	DB = db

	return nil
}

func RunMigrates(db *gorm.DB, models ...any) error {
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			log.Fatalf("failed to migrate model %T: %v", model, err)
		}
	}

	return nil
}

func pgCryptoExtension(db *gorm.DB) error {
	var pgCryptoExtension string = "CREATE EXTENSION IF NOT EXISTS pgcrypto;"
	var checkpgCryptoExtension string = "SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pgcrypto');"

	if err := db.Exec(pgCryptoExtension).Error; err != nil {
		return errors.New("error to enable pgcrypto extension")
	}

	if err := db.Raw(checkpgCryptoExtension).Scan(&exists).Error; err != nil {
		return errors.New("error to check if pgcrypto extension is enabled")
	}

	if !exists {
		return errors.New("pgcrypto extension not found after creation")
	}

	return nil
}

func setPostgresConnectionString() string {
	return "user=" + username + " password=" + password + " host=" + host +
		" port=" + port + " dbname=" + database + " sslmode=" + sslmode + " TimeZone=" + timezone
}
