package producer

import (
	"encoding/json"
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
		var request models.AddItemRequest
		if err := json.Unmarshal(cmdWrapper.Payload, &request); err != nil {
			response, _ := models.SerializeResponse(models.AddItemResponse{
				Success: false,
				Message: "invalid addItem payload",
			})
			return response
		}
		response, _ := models.SerializeResponse(models.AddItemResponse{
			Success: true,
			Message: "item added",
		})
		return response

	case models.DeleteItem:
		var request models.DeleteItemRequest
		if err := json.Unmarshal(cmdWrapper.Payload, &request); err != nil {
			response, _ := models.SerializeResponse(models.DeleteItemResponse{
				Success: false,
				Message: "invalid deleteItem payload",
			})
			return response
		}
		response, _ := models.SerializeResponse(models.DeleteItemResponse{
			Success: true,
			Message: "item deleted",
		})
		return response

	case models.GetItem:
		var request models.GetItemRequest
		if err := json.Unmarshal(cmdWrapper.Payload, &request); err != nil {
			response, _ := models.SerializeResponse(models.GetItemResponse{
				Success: false,
				Message: "invalid getItem payload",
			})
			return response
		}
		response, _ := models.SerializeResponse(models.GetItemResponse{
			Success: true,
			Value:   "testValue1",
		})
		return response

	case models.GetAll:
		var request models.GetAllItemsRequest
		if err := json.Unmarshal(cmdWrapper.Payload, &request); err != nil {
			response, _ := models.SerializeResponse(models.GetAllItemsResponse{
				Success: false,
				Message: "invalid getAllItems payload",
			})
			return response
		}
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

func runProducerTestWithFile(t *testing.T, producer *Producer, fileName string, wg *sync.WaitGroup) {
	defer wg.Done()

	err := producer.ReadCommandsFromFile(fileName)
	assert.NoError(t, err)
}

func responseHandlerDebug(response string) error {
	fmt.Printf("<< Response: %s\n", response)
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
