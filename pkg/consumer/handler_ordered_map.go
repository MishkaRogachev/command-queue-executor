package consumer

import (
	"encoding/json"
	"log"

	"github.com/MishkaRogachev/command-queue-executor/pkg/models"
	"github.com/MishkaRogachev/command-queue-executor/pkg/ordered_map"
)

type RequestHandlerOrderedMap struct {
	omap *ordered_map.OrderedMap[string, string]
}

func NewRequestHandlerOrderedMap() *RequestHandlerOrderedMap {
	return &RequestHandlerOrderedMap{
		omap: ordered_map.New[string, string](),
	}
}

func (h *RequestHandlerOrderedMap) Execute(rawRequest string) string {
	var wrapper models.CommandWrapper
	err := json.Unmarshal([]byte(rawRequest), &wrapper)
	if err != nil {
		log.Printf("Failed to deserialize command wrapper: %v", err)
		return h.errorResponse("invalid command")
	}

	switch wrapper.Type {
	case models.AddItem:
		return h.handleAddItem(wrapper.Payload)
	case models.DeleteItem:
		return h.handleDeleteItem(wrapper.Payload)
	case models.GetItem:
		return h.handleGetItem(wrapper.Payload)
	case models.GetAll:
		return h.handleGetAll(wrapper.Payload)
	default:
		return h.errorResponse("unknown command type")
	}
}

func (h *RequestHandlerOrderedMap) handleAddItem(payload json.RawMessage) string {
	var req models.AddItemRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		log.Printf("Failed to deserialize AddItemRequest: %v", err)
		return h.errorResponse("invalid payload for AddItem")
	}

	h.omap.Store(req.Key, req.Value)
	resp := models.AddItemResponse{
		Success: true,
		Message: "item added",
	}
	return h.toJSON(resp)
}

func (h *RequestHandlerOrderedMap) handleDeleteItem(payload json.RawMessage) string {
	var req models.DeleteItemRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		log.Printf("Failed to deserialize DeleteItemRequest: %v", err)
		return h.errorResponse("invalid payload for DeleteItem")
	}

	err := h.omap.Delete(req.Key)
	if err != nil {
		resp := models.DeleteItemResponse{
			Success: false,
			Message: "key not found",
		}
		return h.toJSON(resp)
	}

	resp := models.DeleteItemResponse{
		Success: true,
		Message: "item deleted",
	}
	return h.toJSON(resp)
}

func (h *RequestHandlerOrderedMap) handleGetItem(payload json.RawMessage) string {
	var req models.GetItemRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		log.Printf("Failed to deserialize GetItemRequest: %v", err)
		return h.errorResponse("invalid payload for GetItem")
	}

	value, err := h.omap.Get(req.Key)
	if err != nil {
		resp := models.GetItemResponse{
			Success: false,
			Message: "key not found",
		}
		return h.toJSON(resp)
	}

	resp := models.GetItemResponse{
		Success: true,
		Value:   value,
	}
	return h.toJSON(resp)
}

func (h *RequestHandlerOrderedMap) handleGetAll(payload json.RawMessage) string {
	var req models.GetAllItemsRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		log.Printf("Failed to deserialize GetAllItemsRequest: %v", err)
		return h.errorResponse("invalid payload for GetAllItems")
	}

	allItems := h.omap.GetAll()
	items := make([]models.KeyValuePair, len(allItems))
	for i, pair := range allItems {
		items[i] = models.KeyValuePair{
			Key:   pair.Key,
			Value: pair.Value,
		}
	}

	resp := models.GetAllItemsResponse{
		Success: true,
		Items:   items,
	}
	return h.toJSON(resp)
}

func (h *RequestHandlerOrderedMap) errorResponse(message string) string {
	return h.toJSON(models.GetAllItemsResponse{
		Success: false,
		Message: message,
	})
}

func (h *RequestHandlerOrderedMap) toJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("Failed to serialize response: %v", err)
		return `{"success":false,"message":"internal error"}`
	}
	return string(data)
}
