package mq

import (
	"errors"
	"sync"
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
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	queue, exists := mq.queues[msg.Topic]
	if !exists {
		return errors.New("topic not subscribed")
	}

	select {
	case queue <- msg:
		return nil
	default:
		return errors.New("queue is full")
	}
}

func (mq *StubMQ) Subscribe(topic string) (<-chan Message, error) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if _, exists := mq.queues[topic]; exists {
		return nil, errors.New("already subscribed to topic")
	}

	queue := make(chan Message, mq.bufferSize)
	mq.queues[topic] = queue
	return queue, nil
}

func (mq *StubMQ) Unsubscribe(topic string) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	queue, exists := mq.queues[topic]
	if !exists {
		return errors.New("not subscribed to topic")
	}

	close(queue)
	delete(mq.queues, topic)
	return nil
}
