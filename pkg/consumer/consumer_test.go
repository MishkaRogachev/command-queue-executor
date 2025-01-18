package consumer

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/MishkaRogachev/command-queue-executor/pkg/mq"
	"github.com/stretchr/testify/assert"
)

// Command generator to produce simple commands
func generateCommands(count int) []string {
	commands := make([]string, count)
	for i := 0; i < count; i++ {
		// TODO: use models!
		commands[i] = fmt.Sprintf(`{"type":"addItem","payload":{"key":"key%d","value":"value%d"}}`, i, i)
	}
	return commands
}

func TestConsumerWithInprocMQ(t *testing.T) {
	// Create an in-process message queue
	inprocMQ := mq.NewInprocMQ()

	// Handler function for the consumer
	handler := func(msg string) string {
		fmt.Printf("Processing message: %s\n", msg)
		return fmt.Sprintf(`{"success":true,"message":"processed %s"}`, msg)
	}

	// Create the consumer
	consumer := NewConsumer(inprocMQ, 3, handler)

	// Start the consumer
	err := consumer.Start()
	assert.NoError(t, err, "failed to start consumer")

	// Generate and publish commands
	commands := generateCommands(10)
	var wg sync.WaitGroup
	for _, cmd := range commands {
		wg.Add(1)
		go func(cmd string) {
			defer wg.Done()
			responseChan, err := inprocMQ.Request(cmd)
			assert.NoError(t, err, "failed to publish command")

			select {
			case response := <-responseChan:
				fmt.Printf("Received response: %s\n", response)
			case <-time.After(2 * time.Second):
				t.Error("Timed out waiting for response")
			}
		}(cmd)
	}

	// Wait for all commands to be processed
	wg.Wait()

	// Stop the consumer
	consumer.Stop()
}
