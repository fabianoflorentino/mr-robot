package app

import (
	"log"

	"github.com/fabianoflorentino/mr-robot/adapters/outbound/persistence/data"
	"github.com/fabianoflorentino/mr-robot/core/domain"
	"github.com/fabianoflorentino/mr-robot/core/services"
	"github.com/fabianoflorentino/mr-robot/database"
	"gorm.io/gorm"
)

type AppContainer struct {
	DB             *gorm.DB
	PaymentService *services.PaymentService
}

func NewAppContainer() (*AppContainer, error) {
	if err := database.InitDB(); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	dbConn := database.DB
	pymtService := paymentService(dbConn)

	return &AppContainer{
		DB:             dbConn,
		PaymentService: pymtService,
	}, nil
}

func paymentService(db *gorm.DB) *services.PaymentService {
	pymt := data.NewDataPaymentRepository(db)
	pymtService := services.NewPaymentService(pymt)

	if err := db.AutoMigrate(&domain.Payment{}); err != nil {
		log.Fatalf("failed to migrate payment repository: %v", err)
	}

	return pymtService
}
