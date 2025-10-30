package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/cache-service/internal/cache"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/cache-service/proto"
)

type Server struct {
	proto.UnimplementedCacheServiceServer
	cache cache.Cache
}

func NewServer(cache cache.Cache) *Server {
	return &Server{cache: cache}
}

func (s *Server) GetMetric(ctx context.Context, req *proto.GetMetricRequest) (*proto.GetMetricResponse, error) {
	metric, err := s.cache.GetMetric(ctx, req.Source, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric from cache: %w", err)
	}

	if metric == nil {
		return &proto.GetMetricResponse{Metric: nil}, nil
	}

	labels := make(map[string]string)
	for k, v := range metric.Labels {
		labels[k] = v.(string)
	}

	return &proto.GetMetricResponse{
		Metric: &proto.Metric{
			Source:      metric.Source,
			Name:        metric.Name,
			Value:       metric.Value,
			Labels:      labels,
			CollectedAt: metric.CollectedAt.Format(time.RFC3339),
		},
	}, nil
}
