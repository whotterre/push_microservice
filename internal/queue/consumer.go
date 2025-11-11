// queue/push_consumer.go
package queue

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/whotterre/push_microservice/internal/dto"
)


type MessageProcessor interface {
	ProcessSendMessage(message []byte) error
	ProcessTokenMessage(message []byte) error
}

type PushConsumer struct {
	conn    *amqp091.Connection
	service MessageProcessor
	workers int
}

func NewPushConsumer(conn *amqp091.Connection, service MessageProcessor, workers int) *PushConsumer {
	return &PushConsumer{
		conn:    conn,
		service: service,
		workers: workers,
	}
}

func (c *PushConsumer) Consume(ctx context.Context) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}

	// Declare queues
	queues := map[string]func(amqp091.Delivery) error{
		"push.send.queue":   c.handleSendMessage,
		"push.tokens.queue": c.handleTokenMessage,
	}

	for queueName, handler := range queues {
		if err := c.setupQueueConsumer(ch, queueName, handler); err != nil {
			_ = ch.Close()
			return err
		}
	}

	<-ctx.Done()
	log.Println("Shutting down consumer...")
	return ch.Close()
}
func (c *PushConsumer) setupQueueConsumer(ch *amqp091.Channel, queueName string, handler func(amqp091.Delivery) error) error {
	// Declare queue (idempotent)
	_, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	// Start consuming
	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	// Process messages with concurrency control
	sem := make(chan struct{}, c.workers)

	go func() {
		for d := range msgs {
			sem <- struct{}{}
			go func(delivery amqp091.Delivery) {
				defer func() { <-sem }()

				start := time.Now()
				correlationID := delivery.CorrelationId
				if correlationID == "" {
					correlationID = "unknown"
				}

				log.Printf("[%s] Processing message from %s", correlationID, queueName)

				if err := handler(delivery); err != nil {
					log.Printf("[%s] Handler failed after %v: %v", correlationID, time.Since(start), err)
					_ = delivery.Nack(false, true)
					return
				}

				log.Printf("[%s] Message processed successfully in %v", correlationID, time.Since(start))
				_ = delivery.Ack(false)
			}(d)
		}
	}()

	log.Printf("Started consumer for queue: %s", queueName)
	return nil
}

func (c *PushConsumer) handleSendMessage(d amqp091.Delivery) error {
	var req dto.PushRequest
	if err := json.Unmarshal(d.Body, &req); err != nil {
		return err
	}

	// Use correlation ID from message if available
	if d.CorrelationId != "" {
		req.CorrelationID = d.CorrelationId
	}

	return c.service.ProcessSendMessage(d.Body)
}

func (c *PushConsumer) handleTokenMessage(d amqp091.Delivery) error {
	var req dto.TokenUpdate
	if err := json.Unmarshal(d.Body, &req); err != nil {
		return err
	}

	return c.service.ProcessTokenMessage(d.Body)
}

func (c *PushConsumer) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}