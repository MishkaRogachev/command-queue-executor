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
	cmdWrapper, err := models.DeserializeCommandWrapper(msg)
	fmt.Println(">> Request:", msg)
	if err != nil {
		response, _ := models.SerializeResponse(models.AddItemResponse{
			Success: false,
			Message: "failed to parse command",
		})
		return response
	}

	switch cmdWrapper.Type {
	case models.AddItem:
		response, _ := models.SerializeResponse(models.AddItemResponse{
			Success: true,
			Message: "item added",
		})
		return response

	case models.DeleteItem:
		response, _ := models.SerializeResponse(models.DeleteItemResponse{
			Success: true,
			Message: "item deleted",
		})
		return response

	case models.GetItem:
		response, _ := models.SerializeResponse(models.GetItemResponse{
			Success: true,
			Value:   "testValue1",
		})
		return response

	case models.GetAll:
		response, _ := models.SerializeResponse(models.GetAllItemsResponse{
			Success: true,
			Items: []models.KeyValuePair{
				{Key: "testKey1", Value: "testValue1"},
			},
		})
		return response

	default:
		response, _ := models.SerializeResponse(models.AddItemResponse{
			Success: false,
			Message: "unknown command",
		})
		return response
	}
}

func runProducerTest(t *testing.T, producer interface{}, wg *sync.WaitGroup) {
	defer wg.Done()

	switch p := producer.(type) {
	case *FileProducer:
		err := p.ReadCommandsFromFile("../../test_data/test_commands_short.txt")
		assert.NoError(t, err)
	case *RandomProducer:
		p.Start()
		time.Sleep(1 * time.Second) // Let the random producer generate commands
		p.Stop()
	}
}

func TestProducersWithInprocMQ(t *testing.T) {
	inprocMQ := mq.NewInprocMQ()

	// Mock server handler
	inprocMQ.ServeHandler(mockServerHandler)

	// Initialize producers
	fileProducer := NewFileProducer(inprocMQ, responseHandlerDebug, 1*time.Second)
	randomProducer := NewRandomProducer(inprocMQ, responseHandlerDebug, 1*time.Second)

	var wg sync.WaitGroup
	wg.Add(2)

	go runProducerTest(t, fileProducer, &wg)
	go runProducerTest(t, randomProducer, &wg)

	wg.Wait()
}

func TestConcurrentProducersWithInprocMQ(t *testing.T) {
	inprocMQ := mq.NewInprocMQ()

	// Mock server handler
	inprocMQ.ServeHandler(mockServerHandler)

	// Initialize producers
	fileProducer1 := NewFileProducer(inprocMQ, responseHandlerDebug, 1*time.Second)
	fileProducer2 := NewFileProducer(inprocMQ, responseHandlerDebug, 1*time.Second)
	randomProducer := NewRandomProducer(inprocMQ, responseHandlerDebug, 1*time.Second)

	var wg sync.WaitGroup
	wg.Add(3)

	go runProducerTest(t, fileProducer1, &wg)
	go runProducerTest(t, fileProducer2, &wg)
	go runProducerTest(t, randomProducer, &wg)

	wg.Wait()
}

func responseHandlerDebug(response string) error {
	fmt.Printf("<< Response: %s\n", response)
	return nil
}
