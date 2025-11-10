package initializers

import (
	amqp091 "github.com/rabbitmq/amqp091-go"
)


func ConnectToRabbitMQ(connString string) (*amqp091.Connection, error){
	var conn *amqp091.Connection
	var err error
	conn, err = amqp091.Dial(connString)
	if err == nil {
		return conn, nil
	}
	return conn, err
}