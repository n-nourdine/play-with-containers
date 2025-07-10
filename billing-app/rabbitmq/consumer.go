// billing
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/n-nourdine/play-with-containers/billing-app/database"
	"github.com/n-nourdine/play-with-containers/billing-app/util"
	"github.com/streadway/amqp"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *log.Logger
	store   *database.OrderStore
}

func NewConsumer(logger *log.Logger, store *database.OrderStore) (*Consumer, error) {
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
		logger.Printf("Tentative de connexion RabbitMQ %d/10 échouée: %v", i+1, err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("impossible de se connecter à RabbitMQ après 10 tentatives: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("impossible d'ouvrir un canal RabbitMQ: %w", err)
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
		return nil, fmt.Errorf("impossible de déclarer la queue: %w", err)
	}

	return &Consumer{
		conn:    conn,
		channel: channel,
		logger:  logger,
		store:   store,
	}, nil
}

func (c *Consumer) StartConsuming(ctx context.Context) error {
	queueName := os.Getenv("RABBITMQ_QUEUE_NAME")
	if queueName == "" {
		queueName = "billing_queue"
	}

	// Set QoS to process one message at a time
	err := c.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("impossible de définir QoS: %w", err)
	}

	messages, err := c.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack (we'll manually ack)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("impossible de commencer la consommation: %w", err)
	}

	c.logger.Printf("En attente de messages sur la queue '%s'. Pour sortir, appuyez sur CTRL+C", queueName)

	go func() {
		for {
			select {
			case msg, ok := <-messages:
				if !ok {
					c.logger.Println("Canal de messages fermé")
					return
				}
				c.processMessage(msg)
			case <-ctx.Done():
				c.logger.Println("Arrêt du consommateur...")
				return
			}
		}
	}()

	return nil
}

func (c *Consumer) processMessage(msg amqp.Delivery) {
	c.logger.Printf("Message reçu: %s", string(msg.Body))

	// Parse the JSON message
	var orderData struct {
		UserID        string `json:"user_id"`
		NumberOfItems string `json:"number_of_items"`
		TotalAmount   string `json:"total_amount"`
	}

	err := json.Unmarshal(msg.Body, &orderData)
	if err != nil {
		c.logger.Printf("Erreur lors du parsing JSON: %v", err)
		// Reject the message without requeue since it's malformed
		msg.Nack(false, false)
		return
	}

	// Validate required fields
	if orderData.UserID == "" || orderData.NumberOfItems == "" || orderData.TotalAmount == "" {
		c.logger.Printf("Champs requis manquants dans le message: %+v", orderData)
		msg.Nack(false, false)
		return
	}

	// Create order with generated ID
	order := database.Order{
		ID:            util.NewUUID(),
		UserID:        orderData.UserID,
		NumberOfItems: orderData.NumberOfItems,
		TotalAmount:   orderData.TotalAmount,
	}

	// Store in database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = c.store.CreateOrder(ctx, order)
	if err != nil {
		c.logger.Printf("Erreur lors de la sauvegarde en base: %v", err)
		// Reject and requeue the message for retry
		msg.Nack(false, true)
		return
	}

	// Acknowledge the message
	err = msg.Ack(false)
	if err != nil {
		c.logger.Printf("Erreur lors de l'accusé de réception: %v", err)
		return
	}

	c.logger.Printf("Commande traitée avec succès: ID=%s, UserID=%s, Items=%s, Total=%s",
		order.ID, order.UserID, order.NumberOfItems, order.TotalAmount)
}

func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
