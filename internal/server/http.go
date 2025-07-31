package server

import (
	"errors"
	"log"
	"os"

	"github.com/fabianoflorentino/mr-robot/adapters/inbound/http/controllers"
	"github.com/fabianoflorentino/mr-robot/internal/app"
	"github.com/gin-gonic/gin"
)

var (
	TRUSTED_PROXIES_ADDRESS []string = []string{"127.0.0.1", "::1", "192.168.0.0/16", "172.16.0.0/8"}
	APP_PORT                string   = os.Getenv("APP_PORT")
)

func InitHTTPServer(container app.Container) {
	g := gin.Default()

	if err := setTrustedProxies(g); err != nil {
		log.Fatalf("failed to set trusted proxies")
	}

	registerPaymentRoutes(g, container)
	registerHealthCheckRoutes(g, container)

	if err := g.Run(":" + APP_PORT); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
}

func registerPaymentRoutes(r *gin.Engine, container app.Container) error {
	paymentController := controllers.NewPaymentController(container.GetPaymentQueue())

	r.POST("/payments", paymentController.ProcessPayment)
	return nil
}

func registerHealthCheckRoutes(r *gin.Engine, container app.Container) error {
	paymentHealthCheck := controllers.NewPaymentController(container.GetPaymentQueue())

	r.GET("/health", paymentHealthCheck.HealthCheck)

	if err := r.SetTrustedProxies(TRUSTED_PROXIES_ADDRESS); err != nil {
		return errors.New("failed to set trusted proxies")
	}

	return nil
}

func setTrustedProxies(e *gin.Engine) error {
	var trustedProxies []string = TRUSTED_PROXIES_ADDRESS

	if err := e.SetTrustedProxies(trustedProxies); err != nil {
		return errors.New("failed to set trusted proxies")
	}

	return nil
}
