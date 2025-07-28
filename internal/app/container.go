package app

import (
	"log"
	"os"

	"github.com/fabianoflorentino/mr-robot/adapters/outbound/gateway"
	"github.com/fabianoflorentino/mr-robot/adapters/outbound/persistence/data"
	"github.com/fabianoflorentino/mr-robot/core/services"
	"github.com/fabianoflorentino/mr-robot/database"
	"github.com/fabianoflorentino/mr-robot/internal/app/queue"
	"gorm.io/gorm"
)

type AppContainer struct {
	DB             *gorm.DB
	PaymentService *services.PaymentService
	PaymentQueue   *queue.PaymentQueue
}

func NewAppContainer() (*AppContainer, error) {
	if err := database.InitDB(); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	dbConn := database.DB
	pymtService := paymentService(dbConn)
	pymtQueue := queue.NewPaymentQueue(4, pymtService)

	return &AppContainer{
		DB:             dbConn,
		PaymentService: pymtService,
		PaymentQueue:   pymtQueue,
	}, nil
}

func paymentService(db *gorm.DB) *services.PaymentService {
	pymt := data.NewDataPaymentRepository(db)
	processor := &gateway.DefaultProcessGateway{
		URL: os.Getenv("DEFAULT_PROCESSOR_URL"),
	}

	pymtService := services.NewPaymentService(pymt, processor)

	if err := db.AutoMigrate(&data.Payment{}); err != nil {
		log.Fatalf("failed to migrate payment repository: %v", err)
	}

	return pymtService
}
