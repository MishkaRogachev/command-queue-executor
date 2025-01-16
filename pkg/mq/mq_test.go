package mq

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func runUnifiedTests(t *testing.T, clientFactory func() ClientMQ, serverFactory func() ServerMQ) {
	t.Run("Single Request-Reply", func(t *testing.T) {
		server := serverFactory()
		defer server.Close()

		err := server.ServeHandler(func(msg string) string {
			return "Reply: " + msg
		})
		assert.NoError(t, err)

		client := clientFactory()
		defer client.Close()

		replyChan, err := client.Request("Test Message")
		assert.NoError(t, err)

		select {
		case reply := <-replyChan:
			assert.Equal(t, "Reply: Test Message", reply)
		case <-time.After(1 * time.Second):
			t.Error("Timed out waiting for reply")
		}
	})

	t.Run("Concurrent Requests", func(t *testing.T) {
		server := serverFactory()
		defer server.Close()

		err := server.ServeHandler(func(msg string) string {
			return fmt.Sprintf("Processed: %s", msg)
		})
		assert.NoError(t, err)

		client := clientFactory()
		defer client.Close()

		var wg sync.WaitGroup
		const concurrentRequests = 10
		for i := 0; i < concurrentRequests; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				message := fmt.Sprintf("Message %d", i)
				replyChan, err := client.Request(message)
				assert.NoError(t, err)

				select {
				case reply := <-replyChan:
					expected := fmt.Sprintf("Processed: %s", message)
					assert.Equal(t, expected, reply)
				case <-time.After(1 * time.Second):
					t.Errorf("Timed out waiting for reply to %s", message)
				}
			}(i)
		}
		wg.Wait()
	})

	t.Run("Multiple Clients", func(t *testing.T) {
		server := serverFactory()
		defer server.Close()

		err := server.ServeHandler(func(msg string) string {
			return "Reply: " + msg
		})
		assert.NoError(t, err)

		const clientCount = 5
		clients := make([]ClientMQ, clientCount)
		for i := 0; i < clientCount; i++ {
			clients[i] = clientFactory()
			defer clients[i].Close()
		}

		var wg sync.WaitGroup
		for i, client := range clients {
			wg.Add(1)
			go func(clientID int, client ClientMQ) {
				defer wg.Done()
				message := fmt.Sprintf("Client %d Message", clientID)
				replyChan, err := client.Request(message)
				assert.NoError(t, err)

				select {
				case reply := <-replyChan:
					expected := "Reply: " + message
					assert.Equal(t, expected, reply)
				case <-time.After(1 * time.Second):
					t.Errorf("Timed out waiting for reply from Client %d", clientID)
				}
			}(i, client)
		}
		wg.Wait()
	})
}

func TestRabbitMQ(t *testing.T) {
	clientFactory := func() ClientMQ {
		client, err := NewClientRabbitMQ("amqp://guest:guest@localhost", 5*time.Second, 3, 1*time.Second)
		if err != nil {
			t.Fatalf("Failed to initialize RabbitMQ client: %v", err)
		}
		return client
	}

	serverFactory := func() ServerMQ {
		server, err := NewServerRabbitMQ("amqp://guest:guest@localhost")
		if err != nil {
			t.Fatalf("Failed to initialize RabbitMQ server: %v", err)
		}
		return server
	}

	runUnifiedTests(t, clientFactory, serverFactory)
}

func TestInprocMQ(t *testing.T) {
	inproc := NewInprocMQ()

	clientFactory := func() ClientMQ {
		return inproc
	}

	serverFactory := func() ServerMQ {
		return inproc
	}

	runUnifiedTests(t, clientFactory, serverFactory)
}
