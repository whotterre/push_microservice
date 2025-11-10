package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/whotterre/push_microservice/internal/config"
	"github.com/whotterre/push_microservice/internal/initializers"
	"github.com/whotterre/push_microservice/internal/routes"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("failed to load config:", err)
		return
	}

	db, err := initializers.ConnectToDB(cfg.PostgresUrl)
	if err != nil {
		return
	}

	err = initializers.PerformMigrations(db)
	if err != nil {
		return
	}
	
	app := fiber.New()
	routes.SetupRoutes(app, cfg, db)

	port := ":" + cfg.Port
	app.Listen(port)
}
