package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/cache-service/proto"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/collector-service/pkg/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CacheClient struct {
	client proto.CacheServiceClient
	conn   *grpc.ClientConn
}

func NewCacheClient(addr string) (*CacheClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC connection: %w", err)
	}
	client := proto.NewCacheServiceClient(conn)
	return &CacheClient{client: client, conn: conn}, nil
}

func (c *CacheClient) GetCachedMetric(ctx context.Context, source, name string) (*models.Metric, error) {
	resp, err := c.client.GetMetric(ctx, &proto.GetMetricRequest{
		Source: source,
		Name:   name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to request metric via gRPC: %w", err)
	}

	if resp.Metric == nil {
		return nil, nil
	}

	labels := make(map[string]any)
	for k, v := range resp.Metric.Labels {
		labels[k] = v
	}

	collectedAt, err := time.Parse(time.RFC3339, resp.Metric.CollectedAt)
	if err != nil {
		return nil, ParseTimeError{
			Field: "collected_at",
			Value: resp.Metric.CollectedAt,
			Err:   err,
		}
	}

	return &models.Metric{
		Source:      resp.Metric.Source,
		Name:        resp.Metric.Name,
		Value:       resp.Metric.Value,
		Labels:      labels,
		CollectedAt: collectedAt,
	}, nil
}

func (c *CacheClient) Close() error {
	return c.conn.Close()
}
