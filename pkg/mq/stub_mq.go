package mq

import (
	"errors"
	"sync"
)

var (
	ErrNoHandler = errors.New("stub-mq: no handler registered")
)

type requestItem struct {
	req        Message
	responseCh chan Message
}

type StubMQ struct {
	requests chan requestItem

	mu      sync.Mutex
	handler func(Message) Message
	closed  bool
}

func NewMessageQueueStub(bufSize int) *StubMQ {
	return &StubMQ{
		requests: make(chan requestItem, bufSize),
	}
}

func (s *StubMQ) Request(msg Message) (<-chan Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, ErrConnectionClosed
	}
	if s.handler == nil {
		return nil, ErrNoHandler
	}

	respCh := make(chan Message, 1)

	s.requests <- requestItem{
		req:        msg,
		responseCh: respCh,
	}

	return respCh, nil
}

func (s *StubMQ) RegisterHandler(handler func(Message) Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrConnectionClosed
	}

	s.handler = handler

	go func() {
		for item := range s.requests {
			response := s.handler(item.req)
			item.responseCh <- response
			close(item.responseCh)
		}
	}()

	return nil
}

func (s *StubMQ) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true
	close(s.requests)
	return nil
}
