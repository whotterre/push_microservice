package handlers

import (
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/whotterre/push_microservice/internal/dto"
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

// SendPush handles synchronous push notification requests
func (h *PushHandler) SendPush(c *fiber.Ctx) error {
	var req dto.PushRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	response, err := h.pushService.SendPushNotification(&req)
	if err != nil {
		log.Printf("Failed to send push notification: %v", err)
		if response != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(response)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to send notification",
		})
	}

	if !response.Success {
		return c.Status(fiber.StatusOK).JSON(response)
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// RegisterDevice registers or updates a device for a user
func (h *PushHandler) RegisterDevice(c *fiber.Ctx) error {
	var req dto.TokenUpdate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Convert to JSON and process through the service
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process request",
		})
	}

	if err := h.pushService.ProcessTokenMessage(reqBytes); err != nil {
		log.Printf("Failed to register device: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to register device",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Device registered successfully",
	})
}

func (h *PushHandler) DoesSomething(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "I am a dummy response",
	})
}

// UpdateNotificationStatus updates the status of a notification
func (h *PushHandler) UpdateNotificationStatus(c *fiber.Ctx) error {
	var req dto.NotificationStatusUpdate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := h.pushService.UpdateNotificationStatus(&req); err != nil {
		log.Printf("Failed to update notification status: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to update notification status",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Notification status updated successfully",
	})
}

// GetNotificationStatus retrieves the status of a notification
func (h *PushHandler) GetNotificationStatus(c *fiber.Ctx) error {
	notificationID := c.Params("notification_id")
	if notificationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "notification_id is required",
		})
	}

	status, err := h.pushService.GetNotificationStatus(notificationID)
	if err != nil {
		log.Printf("Failed to get notification status: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Notification not found",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(status)
}
