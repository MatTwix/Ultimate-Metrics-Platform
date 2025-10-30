package models

import "time"

type Metric struct {
	ID          int64
	Source      string
	Name        string
	Value       float64
	Labels      map[string]any
	CollectedAt time.Time
}
