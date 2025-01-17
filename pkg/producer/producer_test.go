package producer

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/MishkaRogachev/command-queue-executor/pkg/models"
	"github.com/MishkaRogachev/command-queue-executor/pkg/mq"
	"github.com/stretchr/testify/assert"
)

func mockServerHandler(msg string) string {
	cmd, err := models.DeserializeCommandWrapper(msg)
	fmt.Println(">> Received command:", msg)
	if err != nil {
		return `{"success": false, "message": "failed to parse command"}`
	}

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
}

func runProducerTestWithFile(t *testing.T, producer *Producer, fileName string, wg *sync.WaitGroup) {
	defer wg.Done()

	err := producer.ReadCommandsFromFile(fileName)
	assert.NoError(t, err)
}

func responseHandlerDebug(response string) error {
	fmt.Printf("<< Received response: %s\n", response)
	return nil
}

func TestProducerWithInprocMQ(t *testing.T) {
	inprocMQ := mq.NewInprocMQ()

	producer := NewProducer(inprocMQ, responseHandlerDebug, 1*time.Second)

	inprocMQ.ServeHandler(mockServerHandler)

	fileName := "../../test_data/test_commands_short.txt"

	var wg sync.WaitGroup
	wg.Add(1)

	go runProducerTestWithFile(t, producer, fileName, &wg)

	wg.Wait()
}

func TestConcurrentProducersWithInprocMQ(t *testing.T) {
	inprocMQ := mq.NewInprocMQ()

	producer1 := NewProducer(inprocMQ, responseHandlerDebug, 1*time.Second)
	producer2 := NewProducer(inprocMQ, responseHandlerDebug, 1*time.Second)

	inprocMQ.ServeHandler(mockServerHandler)

	fileName1 := "../../test_data/test_commands_medium.txt"
	fileName2 := "../../test_data/test_commands_short.txt"

	var wg sync.WaitGroup
	wg.Add(2)

	go runProducerTestWithFile(t, producer1, fileName1, &wg)
	go runProducerTestWithFile(t, producer2, fileName2, &wg)

	wg.Wait()
}
