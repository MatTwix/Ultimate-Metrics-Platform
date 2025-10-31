package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/api-service/pkg/reader"
	"github.com/MatTwix/Ultimate-Metrics-Platform/services/api-service/proto"
)

type Server struct {
	proto.UnimplementedMetricsServiceServer
	reader reader.MetricsReader
}

func NewServer(reader reader.MetricsReader) *Server {
	return &Server{reader: reader}
}

func (s *Server) GetMetrics(ctx context.Context, req *proto.GetMetricsRequest) (*proto.GetMetricsResponse, error) {
	metrics, err := s.reader.GetMetrics(ctx, req.Source, req.Name, int(req.Limit))
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	protoMetrics := make([]*proto.Metric, 0, len(metrics))

	for _, m := range metrics {
		labels := make(map[string]string)

		for k, v := range m.Labels {
			labels[k] = v.(string)
		}

		protoMetrics = append(protoMetrics, &proto.Metric{
			Source:      m.Source,
			Name:        m.Name,
			Value:       m.Value,
			Labels:      labels,
			CollectedAt: m.CollectedAt.Format(time.RFC3339),
		})
	}

	return &proto.GetMetricsResponse{
		Metrics: protoMetrics,
	}, nil
}

func (s *Server) GetMetric(ctx context.Context, req *proto.GetMetricRequest) (*proto.GetMetricResponse, error) {
	metric, err := s.reader.GetMetric(ctx, req.Source, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric: %w", err)
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
