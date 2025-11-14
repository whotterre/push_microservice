package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rabbitmq/amqp091-go"
	"github.com/whotterre/push_microservice/internal/config"
	"github.com/whotterre/push_microservice/internal/handlers"
	"github.com/whotterre/push_microservice/internal/queue"
	"github.com/whotterre/push_microservice/internal/repository"
	"github.com/whotterre/push_microservice/internal/services"
	"gorm.io/gorm"
)

func SetupRoutes(router *fiber.App, cfg *config.Config, db *gorm.DB, conn *amqp091.Connection, producer queue.PushProducer) *queue.PushConsumer {
	pushRepo := repository.NewPushRepository(db)
	pushService := services.NewPushService(pushRepo, db, conn, producer, cfg)
	pushHandler := handlers.NewPushHandler(pushService)
	consumer := queue.NewPushConsumer(conn, pushService, 10) // 10 workers

	// Production endpoints
	router.Post("/push/send", pushHandler.SendPush)
	router.Post("/push/register", pushHandler.RegisterDevice)
	router.Post("/push/status", pushHandler.UpdateNotificationStatus)
	router.Get("/push/status/:notification_id", pushHandler.GetNotificationStatus)
	router.Put("/push/tokens", pushHandler.DoesSomething) // /push/tokens/{user_id}
	router.Get("/health", pushHandler.GetHealth)

	return consumer
}
