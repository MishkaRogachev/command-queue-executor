package consumer

import (
	"sync"

	"github.com/MishkaRogachev/command-queue-executor/pkg/mq"
)

// RequestHandlerFunc handles a request message and returns a response.
type RequestHandlerFunc func(string) string

// Consumer reads requests from the server, processes them, and replies.
type Consumer struct {
	server      mq.ServerMQ
	handler     RequestHandlerFunc
	workerCount int
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

// NewConsumer creates a new Consumer instance.
func NewConsumer(server mq.ServerMQ, workerCount int, handler RequestHandlerFunc) *Consumer {
	return &Consumer{
		server:      server,
		handler:     handler,
		workerCount: workerCount,
		stopChan:    make(chan struct{}),
	}
}

// Start spawns worker goroutines to process incoming requests.
func (c *Consumer) Start() error {
	reqCh, err := c.server.ListenForRequests()
	if err != nil {
		return err
	}

	for i := 0; i < c.workerCount; i++ {
		c.wg.Add(1)
		go c.worker(reqCh)
	}
	return nil
}

// Stop signals the consumer to stop and waits for workers to finish.
func (c *Consumer) Stop() {
	close(c.stopChan)
	c.wg.Wait()
}

func (c *Consumer) worker(reqCh <-chan mq.Request) {
	defer c.wg.Done()

	for {
		select {
		case req, ok := <-reqCh:
			if !ok {
				// The server closed the requests channel
				return
			}
			response := c.handler(req.Data)
			_ = c.server.Reply(req.CorrelationID, response)
		case <-c.stopChan:
			return
		}
	}
}
