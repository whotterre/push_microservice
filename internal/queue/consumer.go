// queue/push_consumer.go
package queue

import (
	"context"
	"encoding/json"
	"log"
	"strings"
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
		"push.send.queue":        c.handleSendMessage,
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
	// Try to declare queue with current settings
	// If it fails due to existing queue with different args, just consume from existing queue
	_, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // auto-delete
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	if err != nil {
		// If queue already exists with different args, try passive declare
		log.Printf("Queue %s declaration failed: %v. Attempting passive declare...", queueName, err)
		_, err = ch.QueueDeclarePassive(
			queueName,
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Printf("Passive declare also failed. Queue may need manual deletion: %v", err)
			return err
		}
		log.Printf("Using existing queue: %s", queueName)
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

					// Check if error is retryable
					if isRetryableError(err) {
						log.Printf("[%s] Retrying message (transient error)", correlationID)
						// Requeue for retry
						_ = delivery.Nack(false, true)
					} else {
						log.Printf("[%s] Non-retryable error. Rejecting message without requeue.", correlationID)
						// Reject without requeue (message will be discarded)
						_ = delivery.Reject(false)
					}
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
	// Log raw message from queue
	log.Printf("Raw message from queue: %s", string(d.Body))

	var req dto.PushRequest
	if err := json.Unmarshal(d.Body, &req); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return err
	}

	// Log parsed request
	log.Printf("Parsed PushRequest - UserID: %s, Title: '%s', Message: '%s'",
		req.UserID, req.Title, req.Message)

	// Use correlation ID from message if available
	if d.CorrelationId != "" {
		req.CorrelationID = d.CorrelationId
	}

	return c.service.ProcessSendMessage(d.Body)
}

func (c *PushConsumer) handleTokenMessage(d amqp091.Delivery) error {
	// Log raw message from queue
	log.Printf("Raw token update message: %s", string(d.Body))

	var req dto.TokenUpdate
	if err := json.Unmarshal(d.Body, &req); err != nil {
		log.Printf("Failed to unmarshal token update: %v", err)
		return err
	}

	log.Printf("Parsed TokenUpdate - UserID: %s, PlayerID: %s, Platform: %s",
		req.UserID, req.OneSignalPlayerID, req.Platform)

	return c.service.ProcessTokenMessage(d.Body)
}

func (c *PushConsumer) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// Non-retryable errors (business logic failures)
	nonRetryableErrors := []string{
		"no active devices for user",
		"invalid message format",
		"user not found",
		"invalid player id",
	}

	for _, nonRetryable := range nonRetryableErrors {
		if strings.Contains(strings.ToLower(errMsg), nonRetryable) {
			return false
		}
	}

	// Retryable errors (transient failures)
	retryableErrors := []string{
		"connection refused",
		"timeout",
		"temporary failure",
		"database connection",
		"context deadline exceeded",
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(strings.ToLower(errMsg), retryable) {
			return true
		}
	}

	// Default: don't retry unknown errors to avoid infinite loops
	return false
}
