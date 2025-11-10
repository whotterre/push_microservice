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

