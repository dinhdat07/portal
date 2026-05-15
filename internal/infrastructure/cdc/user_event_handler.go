package cdc

import (
	"context"
	"encoding/json"
	"fmt"

	esinfra "portal-system/internal/infrastructure/elasticsearch"
)

type UserEventHandler struct {
	userIndexer *esinfra.UserIndexer
}

func NewUserEventHandler(userIndexer *esinfra.UserIndexer) *UserEventHandler {
	return &UserEventHandler{
		userIndexer: userIndexer,
	}
}

func (h *UserEventHandler) Handle(ctx context.Context, value []byte) error {
	var event DebeziumEvent
	if err := json.Unmarshal(value, &event); err != nil {
		return fmt.Errorf("unmarshal debezium event: %w", err)
	}

	switch event.Op {
	case "r", "c", "u":
		return h.handleUpsert(ctx, event)

	case "d":
		return h.handleDelete(ctx, event)

	default:
		return nil
	}
}

func (h *UserEventHandler) handleUpsert(ctx context.Context, event DebeziumEvent) error {
	if event.After == nil {
		return nil
	}

	doc, err := esinfra.UserDocumentFromDebeziumAfter(event.After)
	if err != nil {
		return fmt.Errorf("map user document: %w", err)
	}

	if doc.ID == "" {
		return nil
	}

	if err := h.userIndexer.Index(ctx, doc); err != nil {
		return fmt.Errorf("index user document: %w", err)
	}

	return nil
}

func (h *UserEventHandler) handleDelete(ctx context.Context, event DebeziumEvent) error {
	if event.Before == nil {
		return nil
	}

	id, ok := event.Before["id"].(string)
	if !ok || id == "" {
		return nil
	}

	if err := h.userIndexer.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user document: %w", err)
	}

	return nil
}
