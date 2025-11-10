package handlers

import (
	"log"

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
	healthResponse, err := h.pushService.GetHealth()
	if err != nil {
		log.Printf("Failed to get health status because: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get health"})
	}

	return c.Status(fiber.StatusOK).JSON(healthResponse)
}

func (h *PushHandler) DoesSomething(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "I am a dummy response",
	})
}
