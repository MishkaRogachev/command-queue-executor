package consumer

import (
	"sync"

	"github.com/MishkaRogachev/command-queue-executor/pkg/mq"
)

type Consumer struct {
	server      mq.ServerMQ
	handler     func(string) string
	workerCount int
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

func NewConsumer(server mq.ServerMQ, workerCount int, handler func(string) string) *Consumer {
	return &Consumer{
		server:      server,
		handler:     handler,
		workerCount: workerCount,
		stopChan:    make(chan struct{}),
	}
}

func (c *Consumer) Start() error {
	err := c.server.ServeHandler(c.processMessage)
	if err != nil {
		return err
	}

	for i := 0; i < c.workerCount; i++ {
		c.wg.Add(1)
		go c.worker()
	}

	return nil
}

func (c *Consumer) Stop() {
	close(c.stopChan)
	c.wg.Wait()
	c.server.Close()
}

func (c *Consumer) worker() {
	defer c.wg.Done()

	<-c.stopChan
}

func (c *Consumer) processMessage(msg string) string {
	return c.handler(msg)
}
