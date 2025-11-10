package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/whotterre/push_microservice/internal/services"
)

type PushHandler struct {
	pushService services.PushService
}

func NewPushHandler(pushService services.PushService) *PushHandler {
	return &PushHandler{
		pushService: pushService,
	}
}


func (h *PushHandler) GetHealth(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"health_status": "healthy", 
	})
}

func (h *PushHandler) DoesSomething(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "I am a dummy response",
	})
}