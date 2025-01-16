package models

import (
	"encoding/json"
	"fmt"
)

type CommandType string

const (
	AddItem    CommandType = "addItem"
	DeleteItem CommandType = "deleteItem"
	GetItem    CommandType = "getItem"
	GetAll     CommandType = "getAllItems"
)

type CommandWrapper struct {
	Type    CommandType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// AddItem Command
type AddItemRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type AddItemResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// DeleteItem Command
type DeleteItemRequest struct {
	Key string `json:"key"`
}

type DeleteItemResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// GetItem Command
type GetItemRequest struct {
	Key string `json:"key"`
}

type GetItemResponse struct {
	Success bool   `json:"success"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message,omitempty"`
}

// GetAllItems Command
type GetAllItemsRequest struct{}

type GetAllItemsResponse struct {
	Items   []KeyValuePair `json:"items"`
	Success bool           `json:"success"`
	Message string         `json:"message,omitempty"`
}

type KeyValuePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func Serialize(commandType CommandType, data any) (string, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to serialize payload: %w", err)
	}

	wrapper := CommandWrapper{
		Type:    commandType,
		Payload: payload,
	}

	raw, err := json.Marshal(wrapper)
	if err != nil {
		return "", fmt.Errorf("failed to serialize command wrapper: %w", err)
	}

	return string(raw), nil
}

func DeserializeWrapper(raw string) (CommandWrapper, error) {
	var wrapper CommandWrapper
	if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
		return CommandWrapper{}, fmt.Errorf("failed to deserialize command wrapper: %w", err)
	}

	return wrapper, nil
}

func Deserialize(raw string, target any) (CommandType, error) {
	wrapper, err := DeserializeWrapper(raw)
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(wrapper.Payload, target); err != nil {
		return "", fmt.Errorf("failed to deserialize payload: %w", err)
	}

	return wrapper.Type, nil
}
