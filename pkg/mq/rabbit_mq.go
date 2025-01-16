package mq

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type MessageRabbitMQ struct {
	Data          string
	ReplyTo       string
	CorrelationID string
}

type ClientRabbitMQ struct {
	conn         *amqp091.Connection
	channel      *amqp091.Channel
	replyQueue   string
	corrMap      sync.Map
	timeout      time.Duration
	retryCount   int
	retryBackoff time.Duration
}

type ServerRabbitMQ struct {
	conn        *amqp091.Connection
	channel     *amqp091.Channel
	handler     func(string) string
	handlerLock sync.Mutex
}

func NewClientRabbitMQ(url string, timeout time.Duration, retryCount int, retryBackoff time.Duration) (*ClientRabbitMQ, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	q, err := ch.QueueDeclare(
		"",    // Name (auto-generated)
		false, // Durable
		true,  // Auto-delete
		true,  // Exclusive
		false, // No-wait
		nil,   // Args
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare reply queue: %w", err)
	}

	client := &ClientRabbitMQ{
		conn:         conn,
		channel:      ch,
		replyQueue:   q.Name,
		timeout:      timeout,
		retryCount:   retryCount,
		retryBackoff: retryBackoff,
	}

	go client.listenForReplies()

	return client, nil
}

func (c *ClientRabbitMQ) listenForReplies() {
	deliveries, err := c.channel.Consume(
		c.replyQueue,
		"",
		true,  // Auto-ack
		true,  // Exclusive
		false, // No local
		false, // No-wait
		nil,   // Args
	)
	if err != nil {
		fmt.Printf("Error starting reply consumer: %v\n", err)
		return
	}

	for d := range deliveries {
		if ch, ok := c.corrMap.Load(d.CorrelationId); ok {
			responseChan := ch.(chan string)
			responseChan <- string(d.Body)
			close(responseChan)
			c.corrMap.Delete(d.CorrelationId)
		}
	}
}

func (c *ClientRabbitMQ) Request(msg string) (<-chan string, error) {
	corrID := uuid.New().String()

	responseChan := make(chan string, 1)
	c.corrMap.Store(corrID, responseChan)

	err := c.channel.Publish(
		"",          // Exchange
		"rpc_queue", // Routing key
		false,
		false,
		amqp091.Publishing{
			ContentType:   "text/plain",
			Body:          []byte(msg),
			ReplyTo:       c.replyQueue,
			CorrelationId: corrID,
		},
	)
	if err != nil {
		c.corrMap.Delete(corrID)
		return nil, fmt.Errorf("failed to publish message: %w", err)
	}

	return responseChan, nil
}

func (c *ClientRabbitMQ) Close() error {
	if err := c.channel.Close(); err != nil {
		return err
	}
	return c.conn.Close()
}

func NewServerRabbitMQ(url string) (*ServerRabbitMQ, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	_, err = ch.QueueDeclare(
		"rpc_queue", // Name
		false,       // Durable
		false,       // Delete when unused
		false,       // Exclusive
		false,       // No-wait
		nil,         // Args
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &ServerRabbitMQ{
		conn:    conn,
		channel: ch,
	}, nil
}

func (s *ServerRabbitMQ) ServeHandler(handler func(string) string) error {
	s.handlerLock.Lock()
	defer s.handlerLock.Unlock()

	if s.handler != nil {
		// Clean up previous handler if it exists
		s.handler = nil
	}

	deliveries, err := s.channel.Consume(
		"rpc_queue",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	s.handler = handler

	go func() {
		for d := range deliveries {
			response := s.handler(string(d.Body))
			err := s.channel.Publish(
				"",
				d.ReplyTo,
				false,
				false,
				amqp091.Publishing{
					ContentType:   "text/plain",
					Body:          []byte(response),
					CorrelationId: d.CorrelationId,
				},
			)
			if err != nil {
				fmt.Printf("failed to send response: %v\n", err)
			}
		}
	}()

	return nil
}

func (s *ServerRabbitMQ) Close() error {
	if err := s.channel.Close(); err != nil {
		return err
	}
	return s.conn.Close()
}
