package dto

import "time"

type GetHealthResponse struct {
	Status       string             `json:"status"`
	Timestamp    time.Time          `json:"timestamp"`
	Service      string             `json:"service"`
	Dependencies DependenciesStatus `json:"dependencies"`
}

type DependenciesStatus struct {
	RabbitMQ        string `json:"rabbitmq"`
	PostgreSQL      string `json:"postgresql"`
}

type PushRequest struct {
	NotificationID string            `json:"notification_id"`
	UserID         string            `json:"user_id"`
	Title          string            `json:"title,omitempty"`
	Message        string            `json:"message,omitempty"`
	TemplateID     string            `json:"template_id,omitempty"`
	TemplateVars   map[string]string `json:"template_variables,omitempty"`
	Data           map[string]interface{} `json:"data,omitempty"`
	Priority       string            `json:"priority,omitempty"` // "high" | "normal"
	CorrelationID  string            `json:"correlation_id"`
}

type TokenUpdate struct {
	UserID             string `json:"user_id"`
	DeviceToken        string `json:"device_token"`
	Platform           string `json:"platform"` // "ios", "android", "web"
	OneSignalPlayerID  string `json:"onesignal_player_id,omitempty"`
}