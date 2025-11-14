package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/whotterre/push_microservice/internal/client"
	"github.com/whotterre/push_microservice/internal/config"
	"github.com/whotterre/push_microservice/internal/dto"
	"github.com/whotterre/push_microservice/internal/models"
	"github.com/whotterre/push_microservice/internal/queue"
	"github.com/whotterre/push_microservice/internal/repository"
	"gorm.io/gorm"
)

type PushService interface {
	GetHealth() (*dto.GetHealthResponse, error)
	ProcessSendMessage(message []byte) error
	ProcessTokenMessage(message []byte) error
	SendPushNotification(req *dto.PushRequest) (*dto.PushResponse, error)
	SendToPlayers(playerIDs []string, title, message string, data map[string]interface{}) (*client.OneSignalResponse, error)
	SendToSegment(segment, title, message string, data map[string]interface{}) (*client.OneSignalResponse, error)
	GetPlayers(limit, offset int) (*client.PlayersResponse, error)
	UpdateNotificationStatus(req *dto.NotificationStatusUpdate) error
	GetNotificationStatus(notificationID string) (*dto.NotificationStatusResponse, error)
}

type pushService struct {
	pushRepo        repository.PushRepository
	bunnyConn       *amqp091.Connection
	db              *gorm.DB
	producer        queue.PushProducer
	oneSignalClient *client.OneSignalClient
}

func NewPushService(pushRepo repository.PushRepository, db *gorm.DB, bunnyConn *amqp091.Connection, producer queue.PushProducer, cfg *config.Config) PushService {
	return &pushService{
		pushRepo:        pushRepo,
		bunnyConn:       bunnyConn,
		db:              db,
		producer:        producer,
		oneSignalClient: client.NewOneSignalClient(cfg),
	}
}

func (s *pushService) ProcessSendMessage(message []byte) error {
	var pushReq dto.PushRequest
	if err := json.Unmarshal(message, &pushReq); err != nil {
		log.Printf("Failed to unmarshal push request: %v", err)
		return fmt.Errorf("invalid message format: %w", err)
	}

	log.Printf("Processing push notification for user: %s", pushReq.UserID)

	// Validate required fields
	if pushReq.Title == "" {
		pushReq.Title = "Notification" // Default title
		log.Printf("Warning: No title provided, using default")
	}
	if pushReq.Message == "" {
		log.Printf("Error: Message is required but was empty")
		return fmt.Errorf("invalid message format: message field is required")
	}

	devices, err := s.pushRepo.GetActiveDevicesByUserID(pushReq.UserID)
	if err != nil {
		log.Printf("Failed to fetch devices for user %s: %v", pushReq.UserID, err)
		return fmt.Errorf("failed to fetch user devices: %w", err)
	}

	if len(devices) == 0 {
		log.Printf("No active devices found for user: %s", pushReq.UserID)
		return fmt.Errorf("no active devices for user: %s", pushReq.UserID)
	}

	playerIDs := make([]string, 0, len(devices))
	for _, device := range devices {
		playerIDs = append(playerIDs, device.PlayerID)
	}

	log.Printf("Sending notification to %d device(s) for user %s. Title: '%s', Message: '%s'",
		len(playerIDs), pushReq.UserID, pushReq.Title, pushReq.Message)

	res, err := s.oneSignalClient.SendToUsers(playerIDs, pushReq.Title, pushReq.Message, pushReq.Data)
	if err != nil {
		log.Printf("Failed to send notification: %v", err)
		return fmt.Errorf("failed to send notification: %w", err)
	}

	log.Printf("Notification sent successfully. ID: %s, Recipients: %d", res.ID, res.Recipients)

	if len(res.GetErrors()) > 0 {
		log.Printf("Notification warnings: %v", res.GetErrors())
	}

	// Create notification log
	notificationLog := &models.NotificationLog{
		NotificationID: res.ID,
		UserID:         pushReq.UserID,
		Status:         string(dto.NotificationStatusPending),
		Recipients:     res.Recipients,
	}
	if err := s.pushRepo.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("Warning: Failed to create notification log: %v", err)
		// Don't fail the request if logging fails
	}

	return nil
}

func (s *pushService) ProcessTokenMessage(message []byte) error {
	var tokenUpdate dto.TokenUpdate
	if err := json.Unmarshal(message, &tokenUpdate); err != nil {
		log.Printf("Failed to unmarshal token update: %v", err)
		return fmt.Errorf("invalid message format: %w", err)
	}

	log.Printf("Processing token update for user: %s, platform: %s", tokenUpdate.UserID, tokenUpdate.Platform)

	// Check if device already exists
	existingDevice, err := s.pushRepo.GetDeviceByPlayerID(tokenUpdate.OneSignalPlayerID)

	if err == nil && existingDevice != nil {
		// Device exists, update it
		existingDevice.UserID = tokenUpdate.UserID
		existingDevice.Platform = tokenUpdate.Platform
		existingDevice.IsActive = true
		if err := s.pushRepo.UpdateDevice(existingDevice); err != nil {
			log.Printf("Failed to update device: %v", err)
			return fmt.Errorf("failed to update device: %w", err)
		}
		log.Printf("Updated existing device for user: %s", tokenUpdate.UserID)
	} else {
		// Create new device
		newDevice := &models.UserDevice{
			UserID:   tokenUpdate.UserID,
			PlayerID: tokenUpdate.OneSignalPlayerID,
			Platform: tokenUpdate.Platform,
			IsActive: true,
		}
		if err := s.pushRepo.CreateDevice(newDevice); err != nil {
			log.Printf("Failed to create device: %v", err)
			return fmt.Errorf("failed to create device: %w", err)
		}
		log.Printf("Created new device for user: %s", tokenUpdate.UserID)
	}

	return nil
}

func (s *pushService) SendPushNotification(req *dto.PushRequest) (*dto.PushResponse, error) {
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if req.Title == "" || req.Message == "" {
		return nil, fmt.Errorf("title and message are required")
	}

	// Get active devices for the user
	devices, err := s.pushRepo.GetActiveDevicesByUserID(req.UserID)
	if err != nil {
		log.Printf("Failed to fetch devices for user %s: %v", req.UserID, err)
		return &dto.PushResponse{
			Success: false,
			Message: "Failed to fetch user devices",
			Errors:  []string{err.Error()},
		}, err
	}

	if len(devices) == 0 {
		log.Printf("No active devices found for user: %s", req.UserID)

		// Create notification log for failed attempt
		errorMsg := fmt.Sprintf("no active devices for user: %s", req.UserID)
		notificationLog := &models.NotificationLog{
			NotificationID: req.NotificationID,
			UserID:         req.UserID,
			Status:         string(dto.NotificationStatusFailed),
			Recipients:     0,
			Error:          &errorMsg,
		}
		if err := s.pushRepo.CreateNotificationLog(notificationLog); err != nil {
			log.Printf("Warning: Failed to create notification log: %v", err)
		}

		return &dto.PushResponse{
			Success: false,
			Message: "No active devices found for user",
		}, nil
	}

	// Extract player IDs
	playerIDs := make([]string, 0, len(devices))
	for _, device := range devices {
		playerIDs = append(playerIDs, device.PlayerID)
	}

	log.Printf("Sending notification to %d device(s) for user %s", len(playerIDs), req.UserID)

	res, err := s.oneSignalClient.SendToUsers(playerIDs, req.Title, req.Message, req.Data)
	if err != nil {
		log.Printf("Failed to send notification: %v", err)

		// Create notification log for failed attempt
		errorMsg := err.Error()
		notificationLog := &models.NotificationLog{
			NotificationID: req.NotificationID,
			UserID:         req.UserID,
			Status:         string(dto.NotificationStatusFailed),
			Recipients:     0,
			Error:          &errorMsg,
		}
		if logErr := s.pushRepo.CreateNotificationLog(notificationLog); logErr != nil {
			log.Printf("Warning: Failed to create notification log: %v", logErr)
		}

		return &dto.PushResponse{
			Success: false,
			Message: "Failed to send notification",
			Errors:  []string{err.Error()},
		}, err
	}

	log.Printf("Notification sent successfully. ID: %s, Recipients: %d", res.ID, res.Recipients)

	// Create notification log
	notificationLog := &models.NotificationLog{
		NotificationID: res.ID,
		UserID:         req.UserID,
		Status:         string(dto.NotificationStatusPending),
		Recipients:     res.Recipients,
	}
	if err := s.pushRepo.CreateNotificationLog(notificationLog); err != nil {
		log.Printf("Warning: Failed to create notification log: %v", err)
		// Don't fail the request if logging fails
	}

	return &dto.PushResponse{
		Success:        true,
		NotificationID: res.ID,
		Recipients:     res.Recipients,
		Errors:         res.GetErrors(),
		Message:        "Notification sent successfully",
	}, nil
}

// SendToPlayers sends a push notification to specific player IDs
func (s *pushService) SendToPlayers(playerIDs []string, title, message string, data map[string]interface{}) (*client.OneSignalResponse, error) {
	return s.oneSignalClient.SendToUsers(playerIDs, title, message, data)
}

// SendToSegment sends a push notification to a segment
func (s *pushService) SendToSegment(segment, title, message string, data map[string]interface{}) (*client.OneSignalResponse, error) {
	return s.oneSignalClient.SendToSegment(segment, title, message, data)
}

// GetPlayers fetches players from OneSignal
func (s *pushService) GetPlayers(limit, offset int) (*client.PlayersResponse, error) {
	return s.oneSignalClient.GetPlayers(limit, offset)
}

func (s *pushService) GetHealth() (*dto.GetHealthResponse, error) {
	// Get RabbitMQ health status
	rabbitStatus := "disconnected"
	if s.bunnyConn != nil && !s.bunnyConn.IsClosed() {
		rabbitStatus = "connected"
	}
	// Get PostgreSQL health status
	postgresStatus := "disconnected"
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err != nil {
			log.Printf("failed to obtain sql.DB from gorm: %v", err)
		} else {
			if err := sqlDB.Ping(); err != nil {
				log.Printf("postgres ping failed: %v", err)
			} else {
				postgresStatus = "connected"
			}
		}
	}

	status := "healthy"
	if rabbitStatus != "connected" && postgresStatus != "connected" {
		status = "unhealthy"
	}

	deps := dto.DependenciesStatus{
		RabbitMQ:   rabbitStatus,
		PostgreSQL: postgresStatus,
	}

	response := dto.GetHealthResponse{
		Status:       status,
		Timestamp:    time.Now().UTC(),
		Service:      "Push Notifications Service",
		Dependencies: deps,
	}

	return &response, nil
}

// UpdateNotificationStatus updates the status of a notification
func (s *pushService) UpdateNotificationStatus(req *dto.NotificationStatusUpdate) error {
	log, err := s.pushRepo.GetNotificationLog(req.NotificationID)
	if err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	// Update status
	log.Status = string(req.Status)
	if req.Error != nil {
		log.Error = req.Error
	}
	log.UpdatedAt = time.Now()

	// Save to database
	if err := s.pushRepo.UpdateNotificationLog(log); err != nil {
		return fmt.Errorf("failed to update notification log: %w", err)
	}

	return nil
}

// GetNotificationStatus retrieves the status of a notification
func (s *pushService) GetNotificationStatus(notificationID string) (*dto.NotificationStatusResponse, error) {
	log, err := s.pushRepo.GetNotificationLog(notificationID)
	if err != nil {
		return nil, fmt.Errorf("notification not found: %w", err)
	}

	response := &dto.NotificationStatusResponse{
		NotificationID: log.NotificationID,
		Status:         dto.NotificationStatus(log.Status),
		Timestamp:      log.UpdatedAt,
		Error:          log.Error,
		UserID:         log.UserID,
		Recipients:     log.Recipients,
	}

	return response, nil
}
