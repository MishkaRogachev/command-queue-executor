package mq

import (
	"errors"
)

// ClientMQ is a message queue client
type ClientMQ interface {
	Request(msg string) (<-chan string, error)
	Close() error
}

// ServerMQ is a message queue server
type ServerMQ interface {
	ServeHandler(handler func(string) string) error
	Close() error
}

var (
	// ErrConnectionClosed is returned when the connection is closed
	ErrConnectionClosed = errors.New("mq: connection is closed")
)
