package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandWrapperSerialization(t *testing.T) {
	t.Run("Serialize and Deserialize AddItemRequest", func(t *testing.T) {
		request := AddItemRequest{
			Key:   "exampleKey",
			Value: "exampleValue",
		}

		raw, err := Serialize(AddItem, request)
		assert.NoError(t, err)
		assert.NotEmpty(t, raw)

		var deserializedRequest AddItemRequest
		commandType, err := Deserialize(raw, &deserializedRequest)
		assert.NoError(t, err)
		assert.Equal(t, AddItem, commandType)
		assert.Equal(t, request, deserializedRequest)
	})

	t.Run("Serialize and Deserialize AddItemResponse", func(t *testing.T) {
		response := AddItemResponse{
			Success: true,
			Message: "Item added successfully",
		}

		raw, err := Serialize(AddItem, response)
		assert.NoError(t, err)
		assert.NotEmpty(t, raw)

		var deserializedResponse AddItemResponse
		commandType, err := Deserialize(raw, &deserializedResponse)
		assert.NoError(t, err)
		assert.Equal(t, AddItem, commandType)
		assert.Equal(t, response, deserializedResponse)
	})

	t.Run("Serialize and Deserialize GetItemRequest", func(t *testing.T) {
		request := GetItemRequest{
			Key: "exampleKey",
		}

		raw, err := Serialize(GetItem, request)
		assert.NoError(t, err)
		assert.NotEmpty(t, raw)

		var deserializedRequest GetItemRequest
		commandType, err := Deserialize(raw, &deserializedRequest)
		assert.NoError(t, err)
		assert.Equal(t, GetItem, commandType)
		assert.Equal(t, request, deserializedRequest)
	})

	t.Run("Serialize and Deserialize GetAllItemsResponse", func(t *testing.T) {
		response := GetAllItemsResponse{
			Items: []KeyValuePair{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
			Success: true,
		}

		raw, err := Serialize(GetAll, response)
		assert.NoError(t, err)
		assert.NotEmpty(t, raw)

		var deserializedResponse GetAllItemsResponse
		commandType, err := Deserialize(raw, &deserializedResponse)
		assert.NoError(t, err)
		assert.Equal(t, GetAll, commandType)
		assert.Equal(t, response, deserializedResponse)
	})
}

func TestInvalidSerialization(t *testing.T) {
	t.Run("Deserialize Invalid JSON", func(t *testing.T) {
		raw := `{"type":"addItem","payload":"{invalid_json"}`
		var request AddItemRequest
		_, err := Deserialize(raw, &request)
		assert.Error(t, err)
	})

	t.Run("Deserialize Unknown Command Type", func(t *testing.T) {
		raw := `{"type":"unknownType","payload":{"key":"exampleKey"}}`
		var request AddItemRequest
		commandType, err := Deserialize(raw, &request)
		assert.NoError(t, err)
		assert.Equal(t, CommandType("unknownType"), commandType)
	})
}
