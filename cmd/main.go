package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Establish a connection to the message queue
	conn, err := initializers.ConnectToRabbitMQ(cfg.RabbitMQURL)
	if err != nil {
		return
	}
	defer conn.Close()

	app := fiber.New()
	consumer := routes.SetupRoutes(app, cfg, db, conn)

	// Start consumer in background with cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := consumer.Consume(ctx); err != nil {
			log.Printf("Consumer error: %v", err)
		}
	}()

	// Start HTTP server in background
	go func() {
		port := ":" + cfg.Port
		log.Printf("Starting server on port %s", port)
		if err := app.Listen(port); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutdown signal received, gracefully stopping...")
	cancel() // stop consumer

	if err := app.Shutdown(); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	time.Sleep(2 * time.Second)
	log.Println("Shutdown complete")
}
