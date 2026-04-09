package main

import (
	"eshkere/internal/app"
	"flag"
	"log"

	"github.com/joho/godotenv"
)

var configPath = flag.String("config", "./config/config.yaml", "path to config file")

// @title           Eshke Reklama API
// @version         1.0
// @description     API для управления рекламодателями и их рекламными кампаниями.
// @host            localhost:8000
// @BasePath        /
// @securityDefinitions.apikey CookieAuth
// @in              cookie
// @name            session_id
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
