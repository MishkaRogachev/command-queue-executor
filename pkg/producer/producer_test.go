package producer

import (
	"fmt"
	"sync"
	"testing"

	"github.com/MishkaRogachev/command-queue-executor/pkg/models"
	"github.com/MishkaRogachev/command-queue-executor/pkg/mq"
	"github.com/stretchr/testify/assert"
)

func TestProducerWithInprocMQ(t *testing.T) {
	inprocMQ := mq.NewInprocMQ()
	producer := NewProducer(inprocMQ)

	// Set up a mock response handler to simulate a server
	inprocMQ.ServeHandler((func(msg string) string {
		cmd, err := models.DeserializeWrapper(msg)
		if err != nil {
			return `{"success": false, "message": "failed to parse command"}`
		}
		fmt.Println("Received command:", cmd.Type)

		switch cmd.Type {
		case models.AddItem:
			return `{"success": true, "message": "item added"}`
		case models.DeleteItem:
			return `{"success": true, "message": "item deleted"}`
		case models.GetItem:
			return `{"success": true, "value": "testValue1"}`
		case models.GetAll:
			return `{"success": true, "items": [{"key": "testKey1", "value": "testValue1"}]}`
		default:
			return `{"success": false, "message": "unknown command"}`
		}
	}))

	fileName := "../../test_data/test_commands_short.txt"

	// Use a WaitGroup to ensure all goroutines complete
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := producer.ReadCommandsFromFile(fileName)
		assert.NoError(t, err)
	}()

	// Wait for all commands to be processed
	wg.Wait()
}
