package elasticsearch

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/testcontainers/testcontainers-go"
	container "github.com/testcontainers/testcontainers-go/modules/elasticsearch"
)

type TestContainer struct {
	Container *container.ElasticsearchContainer
	Config    *Config
}

func StartTestContainer(ctx context.Context) (*TestContainer, error) {
	esContainer, err := container.Run(ctx,
		"docker.elastic.co/elasticsearch/elasticsearch:8.18.8",
		testcontainers.WithEnv(map[string]string{
			"xpack.security.enabled": "false",
			"discovery.type":         "single-node",
			"ES_JAVA_OPTS":           "-Xms256m -Xmx512m",
		}),
	)
	if err != nil || esContainer == nil {
		return nil, fmt.Errorf("failed to start elasticsearch container: %w", err)
	}

	address := esContainer.Settings.Address
	if address == "" {
		return nil, errors.New("failed to get elasticsearch container address")
	}

	cfg := NewConfig([]string{address}, "", "")

	return &TestContainer{
		Container: esContainer,
		Config:    cfg,
	}, nil
}

func (tc *TestContainer) Close(ctx context.Context) error {
	return tc.Container.Terminate(ctx)
}

func (tc *TestContainer) ClearIndices(ctx context.Context, indices ...string) error {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: tc.Config.Addresses,
	})
	if err != nil {
		return fmt.Errorf("failed to create client for clear indices: %w", err)
	}

	target := strings.Join(indices, ",")
	if len(indices) == 0 {
		target = "_all"
	}

	res, err := client.Indices.Delete(
		[]string{target},
		client.Indices.Delete.WithContext(ctx),
		client.Indices.Delete.WithIgnoreUnavailable(true),
	)
	if err != nil {
		return fmt.Errorf("failed to delete indices [%s]: %w", target, err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.IsError() && res.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error deleting indices [%s]: %s", target, res.String())
	}

	return nil
}
