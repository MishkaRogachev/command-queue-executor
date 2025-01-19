package producer

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/MishkaRogachev/command-queue-executor/pkg/models"
	"github.com/MishkaRogachev/command-queue-executor/pkg/mq"
)

// ResponseHandlerFunc is a function type that handles a response message
type ResponseHandlerFunc func(string) error

// RequestFeed is an interface that provides a way to get the next request to be sent to the message queue
type RequestFeed interface {
	Next() (models.RequestWrapper, error)
	IsEmpty() bool
	Close()
}

// Producer responsible for sending requests to the message queue and promoting responses to a handler
type Producer struct {
	client             mq.ClientMQ
	handler            ResponseHandlerFunc
	feed               RequestFeed
	timeout            time.Duration
	maxPendingRequests int
	stopCh             chan struct{}
	wg                 sync.WaitGroup
}

// NewProducer creates a new Producer instance
func NewProducer(
	client mq.ClientMQ,
	handler ResponseHandlerFunc,
	feed RequestFeed,
	timeout time.Duration,
	maxPendingRequests int,
) *Producer {
	return &Producer{
		client:             client,
		handler:            handler,
		timeout:            timeout,
		maxPendingRequests: maxPendingRequests,
		feed:               feed,
		stopCh:             make(chan struct{}),
	}
}

// Start sends requests to the message queue until the request feed is empty
func (p *Producer) Start() {
	pending := make(chan struct{}, p.maxPendingRequests)

	for {
		if p.feed.IsEmpty() {
			break
		}

		select {
		case <-p.stopCh:
			fmt.Println("Producer shutting down gracefully...")
			return
		default:
			request, err := p.feed.Next()
			if err != nil {
				fmt.Printf("Failed to get next request: %v\n", err)
				continue
			}

			p.wg.Add(1)
			pending <- struct{}{}
			go func(req models.RequestWrapper) {
				defer func() {
					p.wg.Done()
					<-pending
				}()

				rawRequest, err := json.Marshal(req)
				if err != nil {
					fmt.Printf("Error serializing request: %v\n", err)
					return
				}

				responseChan, err := p.client.Request(string(rawRequest))
				if err != nil {
					fmt.Printf("Error sending request: %v\n", err)
					return
				}

				select {
				case response := <-responseChan:
					p.handler(response)
				case <-time.After(p.timeout):
					fmt.Println("No response received in time")
				}
			}(request)
		}
	}

	p.wg.Wait()
}

// Close signals the producer to shut down gracefully.
func (p *Producer) Close() {
	close(p.stopCh)
	p.wg.Wait()
	p.feed.Close()
}
