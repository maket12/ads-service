package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	pkgerrs "github.com/maket12/ads-service/backend/authservice/pkg/errs"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/model"
	pkgelasticsearch "github.com/maket12/ads-service/backend/searchservice/pkg/elasticsearch"
)

type esAdDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Category    string `json:"category"`
	MainImage   string `json:"main_image"`
}

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

	defer func() { _ = res.Body.Close() }()

	if res.IsError() {
		return fmt.Errorf("error indexing document ID=%s: %s", adIndex.ID(), res.String())
	}

	return nil
}

func (r *AdIndexRepository) Delete(ctx context.Context, id string) error {
	req := esapi.DeleteRequest{
		Index:      r.indexName,
		DocumentID: id,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.client.Client)
	if err != nil {
		return fmt.Errorf("failed to execute delete request: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			return pkgerrs.NewObjectNotFoundError("ad_index", id)
		}
		return fmt.Errorf("error deleting document ID=%s: %s", id, res.String())
	}

	return nil
}

func (r *AdIndexRepository) Search(ctx context.Context, query *model.SearchQuery) ([]*model.AdIndex, int64, error) {
	mustClause := []map[string]interface{}{}
	filterClause := []map[string]interface{}{}

	if query.Text() != "" {
		mustClause = append(mustClause, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query.Text(),
				"fields": []string{"title^3", "description"},
			},
		})
	} else {
		mustClause = append(mustClause, map[string]interface{}{
			"match_all": map[string]interface{}{},
		})
	}

	if query.Category() != nil {
		filterClause = append(filterClause, map[string]interface{}{
			"term": map[string]interface{}{
				"category": query.Category().String(),
			},
		})
	}

	if query.PriceFrom() != nil || query.PriceTo() != nil {
		priceRange := map[string]interface{}{}
		if query.PriceFrom() != nil {
			priceRange["gte"] = *query.PriceFrom()
		}
		if query.PriceTo() != nil {
			priceRange["lte"] = *query.PriceTo()
		}
		filterClause = append(filterClause, map[string]interface{}{
			"range": map[string]interface{}{
				"price": priceRange,
			},
		})
	}

	var sortQuery []map[string]interface{}
	switch query.SortBy() {
	case model.SortByPriceAsc:
		sortQuery = append(sortQuery, map[string]interface{}{"price": "asc"})
	case model.SortByPriceDesc:
		sortQuery = append(sortQuery, map[string]interface{}{"price": "desc"})
	default:
	}

	esQuery := map[string]interface{}{
		"from": query.Offset(),
		"size": query.Limit(),
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":   mustClause,
				"filter": filterClause,
			},
		},
	}
	if len(sortQuery) > 0 {
		esQuery["sort"] = sortQuery
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(esQuery); err != nil {
		return nil, 0, fmt.Errorf("failed to encode ES query: %w", err)
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(r.indexName),
		r.client.Search.WithBody(&buf),
		r.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute search: %w", err)
	}

	defer func() { _ = res.Body.Close() }()

	if res.IsError() {
		return nil, 0, fmt.Errorf("search response error: %s", res.String())
	}

	var esResponse struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source esAdDTO `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err = json.NewDecoder(res.Body).Decode(&esResponse); err != nil {
		return nil, 0, fmt.Errorf("failed to parse ES response: %w", err)
	}

	resultItems := make([]*model.AdIndex, 0, len(esResponse.Hits.Hits))
	for _, hit := range esResponse.Hits.Hits {
		resultItems = append(resultItems, mapEsDTOToAdIndex(hit.Source))
	}

	return resultItems, esResponse.Hits.Total.Value, nil
}
