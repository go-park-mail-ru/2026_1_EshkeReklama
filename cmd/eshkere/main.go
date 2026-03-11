package main

import (
	"eshkere/internal"
	"flag"
	"log"
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

	application := internal.New(*configPath)
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
