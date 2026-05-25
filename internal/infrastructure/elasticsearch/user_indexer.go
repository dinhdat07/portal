package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	esv8 "github.com/elastic/go-elasticsearch/v8"
)

type UserIndexer struct {
	client *esv8.Client
	index  string
}

func NewUserIndexer(client *esv8.Client, index string) *UserIndexer {
	return &UserIndexer{
		client: client,
		index:  index,
	}
}

func (i *UserIndexer) EnsureIndex(ctx context.Context) error {
	res, err := i.client.Indices.Exists(
		[]string{i.index},
		i.client.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("check index exists: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return nil
	}

	if res.StatusCode != 404 {
		return fmt.Errorf("check index exists failed: %s", res.String())
	}

	mapping := `{
	  "settings": {
	    "analysis": {
	      "normalizer": {
	        "lowercase_normalizer": {
	          "type": "custom",
	          "filter": ["lowercase", "asciifolding"]
	        }
	      }
	    }
	  },
	  "mappings": {
	    "dynamic": false,
	    "properties": {
	      "id": {
	        "type": "keyword"
	      },
	      "email": {
	        "type": "keyword",
	        "normalizer": "lowercase_normalizer"
	      },
	      "username": {
	        "type": "keyword",
	        "normalizer": "lowercase_normalizer"
	      },
	      "first_name": {
	        "type": "text",
	        "fields": {
	          "keyword": {
	            "type": "keyword",
	            "normalizer": "lowercase_normalizer"
	          }
	        }
	      },
	      "last_name": {
	        "type": "text",
	        "fields": {
	          "keyword": {
	            "type": "keyword",
	            "normalizer": "lowercase_normalizer"
	          }
	        }
	      },
	      "full_name": {
	        "type": "text"
	      },
	      "dob": {
	        "type": "date"
	      },
	      "role_id": {
	        "type": "keyword"
	      },
	      "status": {
	        "type": "keyword"
	      },
	      "email_verified_at": {
	        "type": "date"
	      },
	      "last_login_at": {
	        "type": "date"
	      },
	      "created_at": {
	        "type": "date"
	      },
	      "updated_at": {
	        "type": "date"
	      },
	      "deleted_at": {
	        "type": "date"
	      },
	      "deleted_by": {
	        "type": "keyword"
	      }
	    }
	  }
	}`

	createRes, err := i.client.Indices.Create(
		i.index,
		i.client.Indices.Create.WithContext(ctx),
		i.client.Indices.Create.WithBody(strings.NewReader(mapping)),
	)
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		return fmt.Errorf("create index failed: %s", createRes.String())
	}

	return nil
}

func (i *UserIndexer) Index(ctx context.Context, doc UserDocument) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal user document: %w", err)
	}

	res, err := i.client.Index(
		i.index,
		bytes.NewReader(body),
		i.client.Index.WithContext(ctx),
		i.client.Index.WithDocumentID(doc.ID),
		i.client.Index.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("index user document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index user document failed: %s", res.String())
	}

	return nil
}

func (i *UserIndexer) Delete(ctx context.Context, id string) error {
	res, err := i.client.Delete(
		i.index,
		id,
		i.client.Delete.WithContext(ctx),
		i.client.Delete.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("delete user document: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return nil
	}

	if res.IsError() {
		return fmt.Errorf("delete user document failed: %s", res.String())
	}

	return nil
}
