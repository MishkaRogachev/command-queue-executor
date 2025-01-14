package mq

import (
	"errors"
	"sync"
)

const (
	ErrQueueFull         = "queue is full"
	ErrAlreadySubscribed = "already subscribed to topic"
)

type StubMQ struct {
	queues     map[string]chan Message
	bufferSize int
	mu         sync.RWMutex
}

func NewMessageQueueStub(bufferSize int) *StubMQ {
	return &StubMQ{
		queues:     make(map[string]chan Message),
		bufferSize: bufferSize,
	}
}

func (mq *StubMQ) Publish(msg Message) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	queue, exists := mq.queues[msg.Topic]
	if !exists {
		// Create a new channel if no subscription exists
		queue = make(chan Message, mq.bufferSize)
		mq.queues[msg.Topic] = queue
	}

	select {
	case queue <- msg:
		return nil
	default:
		return errors.New(ErrQueueFull)
	}
}

func (mq *StubMQ) Subscribe(topic string) (<-chan Message, error) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	queue, exists := mq.queues[topic]
	if !exists {
		// Create a new channel if no subscription exists
		queue = make(chan Message, mq.bufferSize)
		mq.queues[topic] = queue
	}

	return queue, nil
}

func (mq *StubMQ) Unsubscribe(topic string) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	queue, exists := mq.queues[topic]
	if !exists {
		return ErrNotSubscribedToTopic
	}

	close(queue)
	delete(mq.queues, topic)
	return nil
}
