package services

import (
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/whotterre/push_microservice/internal/dto"
	"github.com/whotterre/push_microservice/internal/queue"
	"github.com/whotterre/push_microservice/internal/repository"
	"gorm.io/gorm"
)

type PushService interface {
	GetHealth() (*dto.GetHealthResponse, error)
	ProcessSendMessage(message []byte) error
	ProcessTokenMessage(message []byte) error
}

type pushService struct {
	pushRepo  repository.PushRepository
	bunnyConn *amqp091.Connection
	db        *gorm.DB
	producer  queue.PushProducer
}

func NewPushService(pushRepo repository.PushRepository, db *gorm.DB, bunnyConn *amqp091.Connection, producer queue.PushProducer) PushService {
	return &pushService{
		pushRepo:  pushRepo,
		bunnyConn: bunnyConn,
		db:        db,
		producer:  producer,
	}
}

func (s *pushService) ProcessSendMessage(message []byte) error {
	return nil
}
func (s *pushService) ProcessTokenMessage(message []byte) error {
	return nil
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
