package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *log.Logger
}

func NewPublisher(logger *log.Logger) (*Publisher, error) {
	rabbitURL := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASSWORD"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)

	// Retry connection with exponential backoff
	var conn *amqp.Connection
	var err error

	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(rabbitURL)
		if err == nil {
			break
		}
		logger.Printf("RabbitMQ connection attempt %d/10 failed: %v", i+1, err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after 10 attempts: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}

	// Declare the queue (idempotent operation)
	queueName := os.Getenv("RABBITMQ_QUEUE_NAME")
	if queueName == "" {
		queueName = "billing_queue"
	}

	_, err = channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	logger.Printf("Connected to RabbitMQ successfully")

	return &Publisher{
		conn:    conn,
		channel: channel,
		logger:  logger,
	}, nil
}

func (p *Publisher) PublishBillingMessage(ctx context.Context, message string) error {
	queueName := os.Getenv("RABBITMQ_QUEUE_NAME")
	if queueName == "" {
		queueName = "billing_queue"
	}

	err := p.channel.Publish(
		"",        // exchange
		queueName, // routing key (queue name)
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // Make message persistent
			Body:         []byte(message),
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.logger.Printf("Message published to queue '%s': %s", queueName, message)
	return nil
}

func (p *Publisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}
