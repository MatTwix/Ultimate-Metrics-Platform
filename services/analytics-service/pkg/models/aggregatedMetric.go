package models

import "time"

type AggregatedMetric struct {
	Source    string
	Name      string
	AvgValue  float64
	MinValue  float64
	MaxValue  float64
	Count     int
	TimeRange string
	StartTime time.Time
	EndTime   time.Time
}
