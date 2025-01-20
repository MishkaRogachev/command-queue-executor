package mq

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

// ClientRabbitMQ implements ClientMQ for RabbitMQ.
type ClientRabbitMQ struct {
	conn       *amqp091.Connection
	pubChannel *amqp091.Channel
	subChannel *amqp091.Channel

	replyQueue string   // ephemeral queue name for this client
	routingKey string   // which queue to publish requests to
	corrMap    sync.Map // correlationID -> chan string
}

// NewClientRabbitMQ creates a new RabbitMQ client with an ephemeral reply queue.
// routingKey is the queue where requests are sent (the "server" queue).
func NewClientRabbitMQ(url, routingKey string) (*ClientRabbitMQ, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Channel #1 for publishing requests
	pubCh, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create publish channel: %w", err)
	}

	// Channel #2 for consuming replies
	subCh, err := conn.Channel()
	if err != nil {
		pubCh.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to create subscribe channel: %w", err)
	}

	// Declare an ephemeral, exclusive queue for this client's replies
	q, err := pubCh.QueueDeclare(
		"",    // name (empty => generated)
		false, // durable
		true,  // auto-delete
		true,  // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		subCh.Close()
		pubCh.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare reply queue: %w", err)
	}

	client := &ClientRabbitMQ{
		conn:       conn,
		pubChannel: pubCh,
		subChannel: subCh,
		replyQueue: q.Name,
		routingKey: routingKey,
		corrMap:    sync.Map{},
	}

	// Start listening for replies in a separate goroutine
	go client.listenForReplies()

	return client, nil
}

// listenForReplies consumes from the ephemeral reply queue and matches messages to correlation IDs.
func (c *ClientRabbitMQ) listenForReplies() {
	deliveries, err := c.subChannel.Consume(
		c.replyQueue,
		"",    // consumer tag
		true,  // auto-ack
		true,  // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		fmt.Printf("Error starting reply consumer: %v\n", err)
		return
	}

	for d := range deliveries {
		corrID := d.CorrelationId
		if ch, ok := c.corrMap.Load(corrID); ok {
			responseChan := ch.(chan string)
			responseChan <- string(d.Body)
			close(responseChan)
			c.corrMap.Delete(corrID)
		}
	}
}

// Request sends `data` to the server queue (routingKey) and returns a channel for the reply.
func (c *ClientRabbitMQ) Request(data string) (<-chan string, error) {
	corrID := uuid.New().String()
	replyChan := make(chan string, 1)

	// Store the channel so listenForReplies can deliver the response
	c.corrMap.Store(corrID, replyChan)

	// Publish the request to the server's queue
	err := c.pubChannel.Publish(
		"",           // exchange (empty => default)
		c.routingKey, // routing key (the queue name)
		false,        // mandatory
		false,        // immediate
		amqp091.Publishing{
			ContentType:   "text/plain",
			Body:          []byte(data),
			CorrelationId: corrID,
			ReplyTo:       c.replyQueue,
		},
	)
	if err != nil {
		c.corrMap.Delete(corrID)
		return nil, fmt.Errorf("failed to publish message: %w", err)
	}

	return replyChan, nil
}

// Close closes channels and connection.
func (c *ClientRabbitMQ) Close() error {
	var firstErr error
	if err := c.pubChannel.Close(); err != nil {
		firstErr = err
	}
	if err := c.subChannel.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := c.conn.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}
