package queue

import (
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

type PushProducer interface {
	PublishMessage(queueName string, message interface{}, correlationID string) error
	PublishToUserService(message interface{}, correlationID string) error
	PublishToEmailService(message interface{}, correlationID string) error
	PublishToTemplateService(message interface{}, correlationID string) error
}

type pushProducer struct {
	conn *amqp091.Connection
}

func NewPushProducer(conn *amqp091.Connection) PushProducer {
	return &pushProducer{
		conn: conn,
	}
}

// Generic method to publish to any queue
func (p *pushProducer) PublishMessage(queueName string, message interface{}, correlationID string) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Declare queue (idempotent)
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	// Marshal message to JSON
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Publish message
	return ch.Publish("", queueName, false, false,
		amqp091.Publishing{
			ContentType:   "application/json",
			Body:          body,
			CorrelationId: correlationID,
			DeliveryMode:  amqp091.Persistent,
		},
	)
}

// Convenience methods for other queues
func (p *pushProducer) PublishToUserService(message interface{}, correlationID string) error {
	return p.PublishMessage("user.send.queue", message, correlationID)
}

func (p *pushProducer) PublishToEmailService(message interface{}, correlationID string) error {
	return p.PublishMessage("user.send.queue", message, correlationID)
}

func (p *pushProducer) PublishToTemplateService(message interface{}, correlationID string) error {
	return p.PublishMessage("template.send.queue", message, correlationID)
}

func (p *pushProducer) Close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
