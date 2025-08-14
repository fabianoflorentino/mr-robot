package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fabianoflorentino/mr-robot/internal/app/container"
	"github.com/fabianoflorentino/mr-robot/internal/server"
)

func main() {
	container := createAppContainer()
	defer gracefulShutdown(container)

	server.InitHTTPServer(container)
}

func createAppContainer() container.Container {
	container, err := container.NewAppContainer()
	if err != nil {
		log.Fatalf("Failed to create app container: %v", err)
	}

	log.Println("Application container initialized successfully")
	return container
}

func gracefulShutdown(container container.Container) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Received shutdown signal, gracefully shutting down...")

		if err := container.Shutdown(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}

		os.Exit(0)
	}()
}
