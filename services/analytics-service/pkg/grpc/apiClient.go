package grpc

import (
	"context"
	"fmt"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/api-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MetricsClient struct {
	client proto.MetricsServiceClient
	conn   *grpc.ClientConn
}

func NewMetricsClient(addr string) (*MetricsClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC connection: %w", err)
	}
	client := proto.NewMetricsServiceClient(conn)
	return &MetricsClient{client: client, conn: conn}, nil
}

func (c *MetricsClient) GetMetrics(ctx context.Context, source, name string, limit int64) ([]*proto.Metric, error) {
	resp, err := c.client.GetMetrics(ctx, &proto.GetMetricsRequest{
		Source: source,
		Name:   name,
		Limit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get via gRPC: %w", err)
	}

	return resp.Metrics, nil
}

func (c *MetricsClient) Close() error {
	return c.conn.Close()
}
