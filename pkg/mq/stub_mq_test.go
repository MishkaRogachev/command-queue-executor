package mq

import (
	"sync"
	"testing"
	"time"
)

func TestMessageQueueStub(t *testing.T) {
	const bufferSize = 10
	mq := NewMessageQueueStub(bufferSize)

	t.Run("Subscribe and Publish", func(t *testing.T) {
		topic := "test-topic"
		msgChan, err := mq.Subscribe(topic)
		if err != nil {
			t.Fatalf("failed to subscribe: %v", err)
		}

		message := Message{Topic: topic, Value: "Hello, Stub!"}
		err = mq.Publish(message)
		if err != nil {
			t.Fatalf("failed to publish message: %v", err)
		}

		select {
		case received := <-msgChan:
			if received != message {
				t.Errorf("expected %v, got %v", message, received)
			}
		case <-time.After(1 * time.Second):
			t.Error("did not receive message in time")
		}
	})

	t.Run("Unsubscribe", func(t *testing.T) {
		topic := "test-topic-unsub"
		msgChan, err := mq.Subscribe(topic)
		if err != nil {
			t.Fatalf("failed to subscribe: %v", err)
		}

		err = mq.Unsubscribe(topic)
		if err != nil {
			t.Fatalf("failed to unsubscribe: %v", err)
		}

		// Verify that the channel is closed
		select {
		case _, ok := <-msgChan:
			if ok {
				t.Error("expected channel to be closed after unsubscribing")
			}
		default:
		}

		err = mq.Publish(Message{Topic: topic, Value: "This should not be received"})
		if err == nil {
			t.Error("expected error when publishing to unsubscribed topic")
		}
	})

	t.Run("Buffer Overflow", func(t *testing.T) {
		topic := "buffer-overflow"
		msgChan, err := mq.Subscribe(topic)
		if err != nil {
			t.Fatalf("failed to subscribe: %v", err)
		}

		for i := 0; i < bufferSize; i++ {
			err := mq.Publish(Message{Topic: topic, Value: "Message"})
			if err != nil {
				t.Fatalf("unexpected error when publishing: %v", err)
			}
		}

		err = mq.Publish(Message{Topic: topic, Value: "Overflow"})
		if err == nil {
			t.Error("expected error when publishing to a full buffer")
		}

		// Drain the channel
		go func() {
			for range msgChan {
			}
		}()
	})

	t.Run("Concurrent Access", func(t *testing.T) {
		topic := "concurrent-topic"
		msgChan, err := mq.Subscribe(topic)
		if err != nil {
			t.Fatalf("failed to subscribe: %v", err)
		}

		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				err := mq.Publish(Message{Topic: topic, Value: "Message"})
				if err != nil {
					t.Errorf("failed to publish in goroutine %d: %v", i, err)
				}
			}(i)
		}

		wg.Wait()

		// Close the test channel
		go func() {
			for range msgChan {
			}
		}()
	})
}
