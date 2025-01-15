package mq

import (
	"errors"
)

type Message struct {
	Data    string
	ReplyTo string
}

type MessageQueue interface {
	Request(msg Message) (<-chan Message, error)
	RegisterHandler(handler func(Message) Message) error
	Close() error
}

var (
	ErrConnectionClosed = errors.New("mq: connection is closed")
)
