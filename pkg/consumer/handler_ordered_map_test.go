package consumer

import (
	"testing"

	"github.com/MishkaRogachev/command-queue-executor/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestRequestHandlerOrderedMapWithModels(t *testing.T) {
	handler := NewRequestHandlerOrderedMap()

	// Test AddItem
	addItemRequest := models.AddItemRequest{Key: "testKey1", Value: "testValue1"}
	rawAddItem, err := models.SerializeRequest(models.AddItem, addItemRequest)
	assert.NoError(t, err)

	response := handler.Execute(rawAddItem)
	var addItemResponse models.AddItemResponse
	err = models.DeserializeResponse(response, &addItemResponse)
	assert.NoError(t, err)
	assert.True(t, addItemResponse.Success)
	assert.Equal(t, "item added", addItemResponse.Message)

	// Test GetItem
	getItemRequest := models.GetItemRequest{Key: "testKey1"}
	rawGetItem, err := models.SerializeRequest(models.GetItem, getItemRequest)
	assert.NoError(t, err)

	response = handler.Execute(rawGetItem)
	var getItemResponse models.GetItemResponse
	err = models.DeserializeResponse(response, &getItemResponse)
	assert.NoError(t, err)
	assert.True(t, getItemResponse.Success)
	assert.Equal(t, "testValue1", getItemResponse.Value)

	// Test DeleteItem
	deleteItemRequest := models.DeleteItemRequest{Key: "testKey1"}
	rawDeleteItem, err := models.SerializeRequest(models.DeleteItem, deleteItemRequest)
	assert.NoError(t, err)

	response = handler.Execute(rawDeleteItem)
	var deleteItemResponse models.DeleteItemResponse
	err = models.DeserializeResponse(response, &deleteItemResponse)
	assert.NoError(t, err)
	assert.True(t, deleteItemResponse.Success)
	assert.Equal(t, "item deleted", deleteItemResponse.Message)

	// Test GetAllItems
	getAllRequest := models.GetAllItemsRequest{}
	rawGetAll, err := models.SerializeRequest(models.GetAll, getAllRequest)
	assert.NoError(t, err)

	response = handler.Execute(rawGetAll)
	var getAllResponse models.GetAllItemsResponse
	err = models.DeserializeResponse(response, &getAllResponse)
	assert.NoError(t, err)
	assert.True(t, getAllResponse.Success)
	assert.Empty(t, getAllResponse.Items)
}
