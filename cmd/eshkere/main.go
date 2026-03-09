package main

import (
	"eshkere/internal/app"
	"flag"
	"log"

	"github.com/joho/godotenv"
)

var configPath = flag.String("config", "./config/config.yaml", "path to config file")

func main() {
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found")
	}

	application := app.New(*configPath)
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
