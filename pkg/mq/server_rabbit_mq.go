package mq

import (
	"fmt"
	"sync"

	"github.com/rabbitmq/amqp091-go"
)

// ServerRabbitMQ implements ServerMQ for RabbitMQ.
type ServerRabbitMQ struct {
	conn       *amqp091.Connection
	channel    *amqp091.Channel
	routingKey string

	requestsCh chan Request
	once       sync.Once

	// correlationID -> replyTo queue
	replyToMap sync.Map
}

// NewServerRabbitMQ creates a server that listens on the named queue (routingKey).
func NewServerRabbitMQ(url, routingKey string) (*ServerRabbitMQ, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Declare the queue the server will consume from (the "server queue").
	_, err = ch.QueueDeclare(
		routingKey,
		false, // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &ServerRabbitMQ{
		conn:       conn,
		channel:    ch,
		routingKey: routingKey,
		requestsCh: make(chan Request),
	}, nil
}

// ListenForRequests returns a channel that the user can read from in worker goroutines.
func (s *ServerRabbitMQ) ListenForRequests() (<-chan Request, error) {
	var initErr error
	s.once.Do(func() {
		deliveries, err := s.channel.Consume(
			s.routingKey,
			"",
			true,  // auto-ack
			false, // not exclusive
			false, // no-local
			false, // no-wait
			nil,
		)
		if err != nil {
			initErr = fmt.Errorf("failed to start consuming: %w", err)
			return
		}

		// Pump deliveries into s.requestsCh
		go func() {
			defer close(s.requestsCh)
			for d := range deliveries {
				// Store the replyTo so we can respond later
				s.replyToMap.Store(d.CorrelationId, d.ReplyTo)

				req := Request{
					Data:          string(d.Body),
					CorrelationID: d.CorrelationId,
					ReplyTo:       d.ReplyTo,
				}
				s.requestsCh <- req
			}
		}()
	})

	if initErr != nil {
		return nil, initErr
	}
	return s.requestsCh, nil
}

// Reply uses correlationID to look up the correct replyTo queue and publishes the response there.
func (s *ServerRabbitMQ) Reply(corrID, data string) error {
	v, ok := s.replyToMap.Load(corrID)
	if !ok {
		return fmt.Errorf("no replyTo found for correlation ID %s", corrID)
	}
	replyTo, _ := v.(string)
	// Optionally delete from the map now to avoid memory leaks
	s.replyToMap.Delete(corrID)

	err := s.channel.Publish(
		"",
		replyTo,
		false,
		false,
		amqp091.Publishing{
			ContentType:   "text/plain",
			Body:          []byte(data),
			CorrelationId: corrID,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}
	return nil
}

// Close closes the RabbitMQ channel and connection.
func (s *ServerRabbitMQ) Close() error {
	if err := s.channel.Close(); err != nil {
		return err
	}
	return s.conn.Close()
}
