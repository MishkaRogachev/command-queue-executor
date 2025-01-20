package consumer

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/MishkaRogachev/command-queue-executor/pkg/mq"
	"github.com/stretchr/testify/assert"
)

// TODO: replace with RandomRequestFeed
func generateCommands(count int) []string {
	commands := make([]string, count)
	for i := 0; i < count; i++ {
		commands[i] = fmt.Sprintf(`{"type":"addItem","payload":{"key":"key%d","value":"value%d"}}`, i, i)
	}
	return commands
}

func TestConsumerWithInproc(t *testing.T) {
	// Create a fresh server
	server := mq.NewInprocServer()
	assert.NotNil(t, server)

	// Create a Consumer with a handler
	handler := func(msg string) string {
		fmt.Printf("Handling: %s\n", msg)
		return fmt.Sprintf(`{"success":true,"message":"processed: %s"}`, msg)
	}
	consumer := NewConsumer(server, 3, handler)
	err := consumer.Start()
	assert.NoError(t, err)

	// Create a client linked to this server
	client := mq.NewInprocClient(server)
	assert.NotNil(t, client)

	// Generate some commands
	commands := generateCommands(10)

	var wg sync.WaitGroup
	for _, cmd := range commands {
		wg.Add(1)
		go func(cmd string) {
			defer wg.Done()

			replyChan, err := client.Request(cmd)
			assert.NoError(t, err)

			select {
			case reply := <-replyChan:
				fmt.Println("Got reply:", reply)
			case <-time.After(2 * time.Second):
				t.Error("Timed out waiting for reply")
			}
		}(cmd)
	}
	wg.Wait()

	consumer.Stop()
	client.Close()
}
