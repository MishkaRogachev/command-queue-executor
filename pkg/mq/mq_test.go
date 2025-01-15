package mq

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func runMessageQueueTests(t *testing.T, mq MessageQueue) {
	t.Run("RegisterHandler", func(t *testing.T) {
		err := mq.RegisterHandler(func(m Message) Message {
			return Message{Data: "echo:" + m.Data}
		})
		assert.NoError(t, err, "failed to register handler")
	})

	t.Run("SingleRequestResponse", func(t *testing.T) {
		req := Message{Data: "test-1"}
		respCh, err := mq.Request(req)
		assert.NoError(t, err, "failed to send request")
		resp := <-respCh
		assert.Equal(t, "echo:test-1", resp.Data, "unexpected response")
	})

	t.Run("ConcurrentRequests", func(t *testing.T) {
		const concurrency = 5
		var wg sync.WaitGroup
		wg.Add(concurrency)

		results := make([]string, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(idx int) {
				defer wg.Done()

				reqMsg := Message{Data: fmt.Sprintf("batch-%d", idx)}
				respCh, err := mq.Request(reqMsg)
				if err != nil {
					t.Errorf("Request error for idx=%d: %v", idx, err)
					return
				}

				select {
				case resp, ok := <-respCh:
					if !ok {
						t.Errorf("Response channel closed unexpectedly for idx %d", idx)
						return
					}
					results[idx] = resp.Data
				case <-time.After(3 * time.Second):
					t.Errorf("Timed out waiting for response at idx %d", idx)
				}
			}(i)
		}
		wg.Wait()

		for i, r := range results {
			expected := fmt.Sprintf("echo:batch-%d", i)
			assert.Equal(t, expected, r, "unexpected concurrent response for idx=%d", i)
		}
	})

	t.Run("CloseAndSend", func(t *testing.T) {
		err := mq.Close()
		assert.NoError(t, err, "failed to close MQ")

		respCh, err2 := mq.Request(Message{Data: "should-fail"})
		assert.Nil(t, respCh, "channel should be nil after close")
		assert.ErrorIs(t, err2, ErrConnectionClosed)
	})
}

func TestStubMQ(t *testing.T) {
	stub := NewMessageQueueStub(10)
	runMessageQueueTests(t, stub)
}

func TestRabbitMQ(t *testing.T) {
	rmq, err := NewRabbitMQ(
		"amqp://guest:guest@localhost:5672/",
		"mq_test_server",
		5*time.Second,        // request timeout
		3,                    // publish retries
		500*time.Millisecond, // backoff
	)
	assert.NoError(t, err, "failed to init RabbitMQ")
	if err != nil {
		t.Skip("Skipping RabbitMQ test; cannot connect.")
	}
	defer func() {
		if cerr := rmq.Close(); cerr != nil {
			log.Printf("failed to close RabbitMQ: %v", cerr)
		}
	}()

	runMessageQueueTests(t, rmq)
}
