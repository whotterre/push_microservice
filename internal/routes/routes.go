package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rabbitmq/amqp091-go"
	"github.com/whotterre/push_microservice/internal/config"
	"github.com/whotterre/push_microservice/internal/handlers"
	"github.com/whotterre/push_microservice/internal/repository"
	"github.com/whotterre/push_microservice/internal/services"
	"gorm.io/gorm"
)

func SetupRoutes(router *fiber.App, cfg *config.Config, db *gorm.DB, conn *amqp091.Connection) {
	pushRepo := repository.NewPushRepository(db)
	pushService := services.NewPushService(pushRepo, db, conn)
	pushHandler := handlers.NewPushHandler(pushService)

	router.Post("/push/send", pushHandler.DoesSomething)
	router.Get("/push/status", pushHandler.DoesSomething) // /push/status/{message_id}
	router.Put("/push/tokens", pushHandler.DoesSomething) // /push/tokens/{user_id}
	router.Get("/health", pushHandler.GetHealth)
}
