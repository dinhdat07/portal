package cdc

import (
	"context"
	"encoding/json"
	"fmt"

	esinfra "portal-system/internal/infrastructure/elasticsearch"
)

type AuditLogEventHandler struct {
	indexer Indexer[esinfra.AuditLogDocument]
}

func NewAuditLogEventHandler(indexer Indexer[esinfra.AuditLogDocument]) *AuditLogEventHandler {
	return &AuditLogEventHandler{
		indexer: indexer,
	}
}

func (h *AuditLogEventHandler) Handle(ctx context.Context, value []byte) error {
	var event DebeziumEvent
	if err := json.Unmarshal(value, &event); err != nil {
		return fmt.Errorf("unmarshal audit log debezium event: %w", err)
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

func (h *AuditLogEventHandler) handleUpsert(ctx context.Context, event DebeziumEvent) error {
	if event.After == nil {
		return nil
	}

	doc := esinfra.AuditLogDocumentFromDebeziumAfter(event.After)
	if doc.ID == "" {
		return nil
	}

	return h.indexer.Index(ctx, doc)
}

func (h *AuditLogEventHandler) handleDelete(ctx context.Context, event DebeziumEvent) error {
	if event.Before == nil {
		return nil
	}

	id, ok := event.Before["id"].(string)
	if !ok || id == "" {
		return nil
	}

	return h.indexer.Delete(ctx, id)
}
