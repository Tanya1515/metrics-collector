package storage

import (
	"context"
)

type RepositoryInterface interface {
	Init(restore bool, fileStore string, backupTime int) error
	// add metric name and value
	RepositoryAddCounterValue(metricName string, metricValue int64) error
	RepositoryAddGaugeValue(metricName string, metricValue float64) error

	// add value
	RepositoryAddValue(metricName string, metricValue int64) error

	// get metric value by name
	GetCounterValueByName(metricName string) (int64, error)
	GetGaugeValueByName(metricName string) (float64, error)

	// check repository availability
	CheckConnection(ctx context.Context) error

	// return all gauge metrics
	GetAllGaugeMetrics() (map[string]float64, error)

	// return all counter metrics
	GetAllCounterMetrics() (map[string]int64, error)
}
