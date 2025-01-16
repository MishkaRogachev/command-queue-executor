package mq

import (
	"errors"
	"sync"
)

type InprocMQ struct {
	mu      sync.RWMutex
	handler func(string) string
}

func NewInprocMQ() *InprocMQ {
	return &InprocMQ{}
}

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

func (mq *InprocMQ) ServeHandler(handler func(string) string) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	mq.handler = handler
	return nil
}

func (mq *InprocMQ) Close() error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	mq.handler = nil
	return nil
}
