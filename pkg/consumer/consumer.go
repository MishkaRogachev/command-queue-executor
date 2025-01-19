package consumer

import (
	"sync"

	"github.com/MishkaRogachev/command-queue-executor/pkg/mq"
)

// RequestHandlerFunc is a function type that handles a request message
// and returns a response message as a string.
type RequestHandlerFunc func(string) string

// Consumer responsible for consuming requests from the message queue and promoting them to a handler
type Consumer struct {
	server      mq.ServerMQ
	handler     RequestHandlerFunc
	workerCount int
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

// NewConsumer creates a new Consumer instance
func NewConsumer(server mq.ServerMQ, workerCount int, handler RequestHandlerFunc) *Consumer {
	return &Consumer{
		server:      server,
		handler:     handler,
		workerCount: workerCount,
		stopChan:    make(chan struct{}),
	}
}

// Start starts the consumer and its worker goroutines
func (c *Consumer) Start() error {
	err := c.server.ServeHandler(c.handler)
	if err != nil {
		return err
	}

	for i := 0; i < c.workerCount; i++ {
		c.wg.Add(1)
		go c.worker()
	}

	return nil
}

// Stop signals the consumer to stop consuming requests
func (c *Consumer) Stop() {
	close(c.stopChan)
	c.wg.Wait()
	c.server.Close()
}

func (c *Consumer) worker() {
	defer c.wg.Done()

	<-c.stopChan
}
