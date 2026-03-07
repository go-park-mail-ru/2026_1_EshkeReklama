package main

import (
	"eshkere/internal/app"
	"flag"
	"log"
)

var configPath = flag.String("config", "./config/config.yaml", "path to config file")

func main() {
	flag.Parse()

	application := app.New(*configPath)
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
