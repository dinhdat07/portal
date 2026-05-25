package elasticsearch

import (
	"context"
	"fmt"
	"log"

	esv8 "github.com/elastic/go-elasticsearch/v8"
)

func Connect(ctx context.Context, address string) (*esv8.Client, error) {
	cfg := esv8.Config{
		Addresses: []string{address},
	}

	es, err := esv8.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create elasticsearch client: %w", err)
	}

	res, err := es.Info(es.Info.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("connect elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch info returned error: %s", res.String())
	}

	log.Println("[Elasticsearch] Connected successfully")

	return es, nil
}
