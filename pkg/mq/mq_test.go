package mq

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func runMessageQueueTests(t *testing.T, clientFactory func() ClientMQ, serverFactory func() ServerMQ) {
	t.Run("Single Request-Reply", func(t *testing.T) {
		server := serverFactory()
		defer server.Close()

		// Start reading requests
		reqCh, err := server.ListenForRequests()
		assert.NoError(t, err)

		// Spin up one goroutine to handle requests
		go func() {
			for req := range reqCh {
				// The server replies with "Reply: <data>"
				_ = server.Reply(req.CorrelationID, "Reply: "+req.Data)
			}
		}()

		// Create a client
		client := clientFactory()
		defer client.Close()

		// Send a request
		replyChan, err := client.Request("Test Message")
		assert.NoError(t, err)

		// Expect "Reply: Test Message"
		select {
		case reply := <-replyChan:
			assert.Equal(t, "Reply: Test Message", reply)
		case <-time.After(time.Second):
			t.Error("Timed out waiting for reply")
		}
	})

	t.Run("Concurrent Requests", func(t *testing.T) {
		server := serverFactory()
		defer server.Close()

		reqCh, err := server.ListenForRequests()
		assert.NoError(t, err)

		// Spin up multiple worker goroutines to handle requests concurrently
		const workerCount = 5
		for i := 0; i < workerCount; i++ {
			go func(workerID int) {
				for req := range reqCh {
					// Always respond with "Reply: <data>"
					_ = server.Reply(req.CorrelationID, "Reply: "+req.Data)
				}
			}(i)
		}

		// Single client that sends multiple concurrent requests
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

				// Expect "Reply: Message N"
				select {
				case reply := <-replyChan:
					expected := "Reply: " + message
					assert.Equal(t, expected, reply)
				case <-time.After(time.Second):
					t.Errorf("Timed out waiting for reply to %s", message)
				}
			}(i)
		}
		wg.Wait()
	})

	t.Run("Multiple Clients", func(t *testing.T) {
		server := serverFactory()
		defer server.Close()

		// Start reading requests in a goroutine
		reqCh, err := server.ListenForRequests()
		assert.NoError(t, err)

		go func() {
			for req := range reqCh {
				// Reply with "Reply: <data>"
				_ = server.Reply(req.CorrelationID, "Reply: "+req.Data)
			}
		}()

		// Create multiple clients
		const clientCount = 5
		clients := make([]ClientMQ, clientCount)
		for i := 0; i < clientCount; i++ {
			clients[i] = clientFactory()
			// Defer each close
			defer clients[i].Close()
		}

		// Each client sends a request concurrently
		var wg sync.WaitGroup
		for i, c := range clients {
			wg.Add(1)
			go func(clientID int, cl ClientMQ) {
				defer wg.Done()

				message := fmt.Sprintf("Client %d Message", clientID)
				replyChan, err := cl.Request(message)
				assert.NoError(t, err)

				// Expect "Reply: Client N Message"
				select {
				case reply := <-replyChan:
					expected := "Reply: " + message
					assert.Equal(t, expected, reply)
				case <-time.After(time.Second):
					t.Errorf("Timed out waiting for reply from Client %d", clientID)
				}
			}(i, c)
		}
		wg.Wait()
	})
}

func TestRabbitMQ(t *testing.T) {
	serverFactory := func() ServerMQ {
		server, err := NewServerRabbitMQ(GetRabbitMQURL(), "test-exchange")
		if err != nil {
			t.Fatalf("Failed to initialize RabbitMQ server: %v", err)
		}
		return server
	}

	clientFactory := func() ClientMQ {
		client, err := NewClientRabbitMQ(GetRabbitMQURL(), "test-exchange")
		if err != nil {
			t.Fatalf("Failed to initialize RabbitMQ client: %v", err)
		}
		return client
	}

	runMessageQueueTests(t, clientFactory, serverFactory)
}

func TestInprocMQ(t *testing.T) {
	var inprocServer *InprocServer

	serverFactory := func() ServerMQ {
		inprocServer = NewInprocServer()
		return inprocServer
	}

	clientFactory := func() ClientMQ {
		return NewInprocClient(inprocServer) // link to the current server
	}

	runMessageQueueTests(t, clientFactory, serverFactory)
}
