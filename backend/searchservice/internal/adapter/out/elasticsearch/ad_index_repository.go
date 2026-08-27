package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/model"
	pkgelasticsearch "github.com/maket12/ads-service/backend/searchservice/pkg/elasticsearch"
)

type AdIndexRepository struct {
	client    *pkgelasticsearch.Client
	indexName string
}

func NewAdIndexRepository(client *pkgelasticsearch.Client, indexName string) *AdIndexRepository {
	return &AdIndexRepository{
		client:    client,
		indexName: indexName,
	}
}

func (r *AdIndexRepository) Index(ctx context.Context, adIndex *model.AdIndex) error {
	data, err := json.Marshal(mapAdIndexToEsDTO(adIndex))
	if err != nil {
		return fmt.Errorf("failed to marshal ad index document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      r.indexName,
		DocumentID: adIndex.ID(),
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.client.Client)
	if err != nil {
		return fmt.Errorf("failed to execute index request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document ID=%s: %s", adIndex.ID(), res.String())
	}

	return nil
}

func
