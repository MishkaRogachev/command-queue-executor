package mq

import (
	"errors"
)

type ClientMQ interface {
	Request(msg string) (<-chan string, error)
	Close() error
}

type ServerMQ interface {
	ServeHandler(handler func(string) string) error
	Close() error
}

var (
	ErrConnectionClosed = errors.New("mq: connection is closed")
)
