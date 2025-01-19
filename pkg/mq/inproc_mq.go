package mq

import (
	"errors"
	"sync"
)

// InprocMQ is an in-process message queue
type InprocMQ struct {
	mu      sync.RWMutex
	handler func(string) string
}

// NewInprocMQ creates a new InprocMQ instance
func NewInprocMQ() *InprocMQ {
	return &InprocMQ{}
}

// Request sends a message to the message queue and returns a channel to receive the response
func (mq *InprocMQ) Request(msg string) (<-chan string, error) {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	if mq.handler == nil {
		return nil, errors.New("no handler registered")
	}

	replyChan := make(chan string, 1)
	go func() {
		defer close(replyChan)
		response := mq.handler(msg)
		replyChan <- response
	}()
	return replyChan, nil
}

// ServeHandler registers a handler function to process messages
func (mq *InprocMQ) ServeHandler(handler func(string) string) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	mq.handler = handler
	return nil
}

// Close closes the message queue
func (mq *InprocMQ) Close() error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	mq.handler = nil
	return nil
}
