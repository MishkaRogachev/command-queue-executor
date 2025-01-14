package mq

import "errors"

var (
	ErrNotSubscribedToTopic = errors.New("not subscribed to topic")
)

type Message struct {
	Topic string
	Value string
}

type MessageQueue interface {
	Publish(msg Message) error
	Subscribe(topic string) (<-chan Message, error)
	Unsubscribe(topic string) error
}
