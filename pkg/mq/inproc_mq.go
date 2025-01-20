package mq

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// InprocServer is an in-process message queue server
type InprocServer struct {
	mu            sync.RWMutex
	requests      chan Request
	corrClientMap sync.Map
}

// NewInprocServer is an in-process message queue client
func NewInprocServer() *InprocServer {
	return &InprocServer{
		requests: make(chan Request),
	}
}

// ListenForRequests returns a channel that the user can read from in worker goroutines
func (s *InprocServer) ListenForRequests() (<-chan Request, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.requests, nil
}

// Reply sends a reply message to the client
func (s *InprocServer) Reply(corrID, data string) error {
	v, ok := s.corrClientMap.Load(corrID)
	if !ok {
		return fmt.Errorf("no client for correlation ID %s", corrID)
	}
	client := v.(*InprocClient)
	return client.deliverReply(corrID, data)
}

// Close closes the server
func (s *InprocServer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	close(s.requests)
	return nil
}

func (s *InprocServer) acceptRequest(r Request, c *InprocClient) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.corrClientMap.Store(r.CorrelationID, c)
	s.requests <- r
	return nil
}

// InprocClient is an in-process message queue client
type InprocClient struct {
	mu      sync.RWMutex
	server  *InprocServer
	corrMap sync.Map
}

// NewInprocClient creates a new in-process client
func NewInprocClient(server *InprocServer) *InprocClient {
	return &InprocClient{server: server}
}

// Request sends a request message to the server
func (c *InprocClient) Request(data string) (<-chan string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	corrID := uuid.New().String()
	replyChan := make(chan string, 1)
	c.corrMap.Store(corrID, replyChan)
	req := Request{Data: data, CorrelationID: corrID, ReplyTo: corrID}
	if err := c.server.acceptRequest(req, c); err != nil {
		c.corrMap.Delete(corrID)
		return nil, err
	}
	return replyChan, nil
}

// Close closes the client
func (c *InprocClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return nil
}

func (c *InprocClient) deliverReply(corrID, data string) error {
	v, ok := c.corrMap.Load(corrID)
	if !ok {
		return nil
	}
	replyChan := v.(chan string)
	replyChan <- data
	close(replyChan)
	c.corrMap.Delete(corrID)
	return nil
}
