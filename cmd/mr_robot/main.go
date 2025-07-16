package main

import (
	"fmt"
	"os"

	"github.com/fabianoflorentino/mr-robot/config"
)

func init() {
	config.LoadEnv()
}

func main() {
	var app_name string = os.Getenv("APP_NAME")
	fmt.Println("Hello " + app_name + "!")
}
