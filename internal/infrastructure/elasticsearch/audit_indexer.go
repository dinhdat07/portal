package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	esv8 "github.com/elastic/go-elasticsearch/v8"
)

type AuditLogIndexer struct {
	client *esv8.Client
	index  string
}

func NewAuditLogIndexer(client *esv8.Client, index string) *AuditLogIndexer {
	return &AuditLogIndexer{
		client: client,
		index:  index,
	}
}

func (i *AuditLogIndexer) EnsureIndex(ctx context.Context) error {
	res, err := i.client.Indices.Exists(
		[]string{i.index},
		i.client.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("check audit log index exists: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return nil
	}

	if res.StatusCode != 404 {
		return fmt.Errorf("check audit log index exists failed: %s", res.String())
	}

	mapping := `{
	  "mappings": {
	    "dynamic": false,
	    "properties": {
	      "id": { "type": "keyword" },
	      "action": { "type": "keyword" },

	      "actor_user_id": { "type": "keyword" },
	      "actor_username": {
	        "type": "text",
	        "fields": {
	          "keyword": { "type": "keyword" }
	        }
	      },
	      "actor_email": { "type": "keyword" },
	      "actor_role": { "type": "keyword" },

	      "target_user_id": { "type": "keyword" },
	      "target_username": {
	        "type": "text",
	        "fields": {
	          "keyword": { "type": "keyword" }
	        }
	      },
	      "target_email": { "type": "keyword" },
	      "target_role": { "type": "keyword" },

	      "metadata": {
	        "type": "object",
	        "enabled": true
	      },
	      "ip_address": { "type": "ip" },
	      "user_agent": { "type": "text" },
	      "created_at": { "type": "date" }
	    }
	  }
	}`

	createRes, err := i.client.Indices.Create(
		i.index,
		i.client.Indices.Create.WithContext(ctx),
		i.client.Indices.Create.WithBody(strings.NewReader(mapping)),
	)
	if err != nil {
		return fmt.Errorf("create audit log index: %w", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		return fmt.Errorf("create audit log index failed: %s", createRes.String())
	}

	return nil
}

func (i *AuditLogIndexer) Index(ctx context.Context, doc AuditLogDocument) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal audit log document: %w", err)
	}

	res, err := i.client.Index(
		i.index,
		bytes.NewReader(body),
		i.client.Index.WithContext(ctx),
		i.client.Index.WithDocumentID(doc.ID),
		i.client.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("index audit log document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index audit log document failed: %s", res.String())
	}

	return nil
}

func (i *AuditLogIndexer) Delete(ctx context.Context, id string) error {
	res, err := i.client.Delete(
		i.index,
		id,
		i.client.Delete.WithContext(ctx),
		i.client.Delete.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("delete audit log document: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return nil
	}

	if res.IsError() {
		return fmt.Errorf("delete audit log document failed: %s", res.String())
	}

	return nil
}
