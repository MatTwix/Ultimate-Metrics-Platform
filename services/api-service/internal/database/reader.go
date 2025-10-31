package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/MatTwix/Ultimate-Metrics-Platform/api-service/pkg/models"
	"github.com/MatTwix/Ultimate-Metrics-Platform/api-service/pkg/reader"
)

type PostgresMetricsReader struct {
	storage *Storage
}

func NewPostgresMetricsReeader(storage *Storage) reader.MetricsReader {
	return &PostgresMetricsReader{storage: storage}
}

func (r *PostgresMetricsReader) GetMetrics(ctx context.Context, source, name string, limit int) ([]models.Metric, error) {
	query := `SELECT source, name, value, labels, collected_at FROM metrics WHERE source = $1 AND name = $2 ORDER BY collected_at DESC LIMIT $3`
	rows, err := r.storage.db.QueryContext(ctx, query, source, name, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to run query '%s': %w", query, err)
	}
	defer rows.Close()

	var metrics []models.Metric
	var mulSelErr MultipleSelectError
	for rows.Next() {
		var m models.Metric
		var labelsJSON []byte
		err := rows.Scan(&m.Source, &m.Name, &m.Value, &labelsJSON, &m.CollectedAt)
		if err != nil {
			mulSelErr.Errors = append(mulSelErr.Errors, fmt.Errorf("failed to scan row: %w", err))
			mulSelErr.FailedCount++
			continue
		}

		if err := json.Unmarshal(labelsJSON, &m.Labels); err != nil {
			mulSelErr.Errors = append(mulSelErr.Errors, fmt.Errorf("failed to unmarshal labels '%v': %w", labelsJSON, err))
			mulSelErr.FailedCount++
			continue
		}

		metrics = append(metrics, m)
		mulSelErr.SuccessfullCount++
	}

	if len(mulSelErr.Errors) > 0 {
		return metrics, &mulSelErr
	}

	return metrics, nil
}

func (r *PostgresMetricsReader) GetMetric(ctx context.Context, source, name string) (*models.Metric, error) {
	query := `SELECT source, name, value, labels, collected_at FROM metrics WHERE source = $1 AND name = $2 ORDER BY collected_at DESC LIMIT 1`
	var m models.Metric
	var labelsJSON []byte

	if err := r.storage.db.QueryRowContext(ctx, query, source, name).Scan(
		&m.Source,
		&m.Name,
		&m.Value,
		&labelsJSON,
		&m.CollectedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	if err := json.Unmarshal(labelsJSON, &m.Labels); err != nil {
		return nil, fmt.Errorf("failed to unmarshal labels '%v': %w", labelsJSON, err)
	}

	return &m, nil
}
