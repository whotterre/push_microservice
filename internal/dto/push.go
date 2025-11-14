package dto

import "time"

type GetHealthResponse struct {
	Status       string             `json:"status"`
	Timestamp    time.Time          `json:"timestamp"`
	Service      string             `json:"service"`
	Dependencies DependenciesStatus `json:"dependencies"`
}

type DependenciesStatus struct {
	RabbitMQ   string `json:"rabbitmq"`
	PostgreSQL string `json:"postgresql"`
}

type PushRequest struct {
	NotificationID string                 `json:"notification_id"`
	UserID         string                 `json:"user_id"`
	Title          string                 `json:"title,omitempty"`
	Message        string                 `json:"message,omitempty"`
	TemplateID     string                 `json:"template_id,omitempty"`
	TemplateVars   map[string]string      `json:"template_variables,omitempty"`
	Data           map[string]interface{} `json:"data,omitempty"`
	Priority       string                 `json:"priority,omitempty"` // "high" | "normal"
	CorrelationID  string                 `json:"correlation_id"`
}

type TokenUpdate struct {
	UserID            string `json:"user_id"`
	DeviceToken       string `json:"device_token"`
	Platform          string `json:"platform"` // "ios", "android", "web"
	OneSignalPlayerID string `json:"onesignal_player_id,omitempty"`
}

type PushResponse struct {
	Success        bool     `json:"success"`
	NotificationID string   `json:"notification_id,omitempty"`
	Recipients     int      `json:"recipients"`
	Errors         []string `json:"errors,omitempty"`
	Message        string   `json:"message,omitempty"`
}

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusFailed    NotificationStatus = "failed"
)

// NotificationStatusUpdate represents a status update for a notification
type NotificationStatusUpdate struct {
	NotificationID string             `json:"notification_id" validate:"required"`
	Status         NotificationStatus `json:"status" validate:"required"`
	Timestamp      *time.Time         `json:"timestamp,omitempty"`
	Error          *string            `json:"error,omitempty"`
}

// NotificationStatusResponse represents the response for status queries
type NotificationStatusResponse struct {
	NotificationID string             `json:"notification_id"`
	Status         NotificationStatus `json:"status"`
	Timestamp      time.Time          `json:"timestamp"`
	Error          *string            `json:"error,omitempty"`
	UserID         string             `json:"user_id,omitempty"`
	Recipients     int                `json:"recipients,omitempty"`
}
