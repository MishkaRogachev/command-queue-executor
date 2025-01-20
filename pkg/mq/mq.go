package mq

import (
	"os"
)

// Request is the server's view of an incoming message: what data arrived,
// plus correlation info so we can reply back to the correct client.
type Request struct {
	Data          string
	CorrelationID string
	ReplyTo       string
}

// ClientMQ is the interface for any message queue client implementation.
type ClientMQ interface {
	// Request sends the given data as a request and returns a channel to receive the reply.
	Request(data string) (<-chan string, error)
	// Close should close any underlying network connections, channels, etc.
	Close() error
}

// ServerMQ is the interface for any message queue server implementation.
type ServerMQ interface {
	// ListenForRequests returns a channel on which the server can read incoming requests.
	ListenForRequests() (<-chan Request, error)
	// Reply allows the server to send a response for the given correlation ID.
	Reply(corrID, data string) error
	// Close closes underlying resources like channels/connections.
	Close() error
}

// GetRabbitMQURL returns the RabbitMQ URL from the environment (RABBITMQ_URL) or a default.
func GetRabbitMQURL() string {
	if url := os.Getenv("RABBITMQ_URL"); url != "" {
		return url
	}
	return "amqp://guest:guest@localhost"
}
