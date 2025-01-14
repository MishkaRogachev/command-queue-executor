package mq

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func runMessageQueueTests(t *testing.T, mq MessageQueue) {
	t.Run("Subscribe and Publish", func(t *testing.T) {
		topic := "test-topic"
		msgChan, err := mq.Subscribe(topic)
		assert.NoError(t, err, "failed to subscribe")

		message := Message{Topic: topic, Value: "Hello, MessageQueue!"}
		err = mq.Publish(message)
		assert.NoError(t, err, "failed to publish message")

		select {
		case received := <-msgChan:
			assert.Equal(t, message, received, "received message does not match")
		case <-time.After(1 * time.Second):
			t.Error("did not receive message in time")
		}
	})

	t.Run("Unsubscribe", func(t *testing.T) {
		topic := "test-topic-unsub"
		msgChan, err := mq.Subscribe(topic)
		assert.NoError(t, err, "failed to subscribe")

		err = mq.Unsubscribe(topic)
		assert.NoError(t, err, "failed to unsubscribe")

		// Verify that the channel is closed
		select {
		case _, ok := <-msgChan:
			assert.False(t, ok, "expected channel to be closed after unsubscribing")
		default:
		}
	})

	t.Run("Concurrent Access", func(t *testing.T) {
		topic := "concurrent-topic"
		msgChan, err := mq.Subscribe(topic)
		assert.NoError(t, err, "failed to subscribe")

		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				err := mq.Publish(Message{Topic: topic, Value: "Message"})
				assert.NoError(t, err, "failed to publish in goroutine")
			}(i)
		}

		wg.Wait()

		// Drain the channel
		go func() {
			for range msgChan {
			}
		}()
	})
}

func TestMessageQueueStub(t *testing.T) {
	mq := NewMessageQueueStub(10) // Buffer size for stub
	runMessageQueueTests(t, mq)
}

func TestRabbitMQ(t *testing.T) {
	rabbitMQ, err := NewRabbitMQ(
		"amqp://guest:guest@localhost",
		"events",
	)
	assert.NoError(t, err, "failed to initialize RabbitMQ")
	defer func() {
		if err := rabbitMQ.Close(); err != nil {
			log.Printf("failed to close RabbitMQ: %v", err)
		}
	}()

	runMessageQueueTests(t, rabbitMQ)
}
