package main

import (
	"log"

	"github.com/fabianoflorentino/mr-robot/config"
	"github.com/fabianoflorentino/mr-robot/internal/app"
	"github.com/fabianoflorentino/mr-robot/internal/server"
)

func init() {
	config.LoadEnv()
}

func main() {
	c := appContainer()
	server.InitHTTPServer(c)
}

func appContainer() *app.AppContainer {
	container, err := app.NewAppContainer()
	if err != nil {
		log.Fatalf("error to instace a new app container: %v", err)
	}

	return container
}
