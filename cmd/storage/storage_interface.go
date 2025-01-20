package storage

import (
	"context"
)

type RepositoryInterface interface {

	Init() (error)
	// add metric name and value 
	RepositoryAddCounterValue(metricName string, metricValue int64)
	RepositoryAddGaugeValue(metricName string, metricValue float64)

	// add value 
	RepositoryAddValue(metricName string, metricValue int64)

	// get metric value by name
	GetCounterValueByName(metricName string) (int64, error)
	GetGaugeValueByName(metricName string) (float64, error)

	// check repository availability
	CheckConnection(ctx context.Context) (error)

	// return all gauge metrics
	GetAllGaugeMetrics() (map[string]float64)

	// return all counter metrics
	GetAllCounterMetrics() (map[string]int64)
}
