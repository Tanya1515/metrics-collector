// Storage implements interface fo saving metrics in two
// ways: to PostgreSQL and to in-memory storage (MemStorage)
// Historical data can be retrieved from back-up file
package storage

import (
	"context"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
)

type RepositoryInterface interface {
	// Init - function for initialization in-memoty/PostgreSQL storage.
	Init(ctx context.Context, shutdown chan struct{}) error

	// RepositoryAddCounterValue - function for modifying/adding new counter metric in PostgreSQL/in-memory storage.
	RepositoryAddCounterValue(metricName string, metricValue int64) error

	// RepositoryAddGaugeValue - function for modifying/adding new gauge metric to PostgreSQL/in-memory storage.
	RepositoryAddGaugeValue(metricName string, metricValue float64) error

	// RepositoryAddValue - function for modifying/adding new metric to PostgreSQL/in-memory storage.
	RepositoryAddValue(metricName string, metricValue int64) error

	// GetCounterValueByName - function for getting counter metric value by it's name from PostgreSQL/in-memory storage.
	GetCounterValueByName(metricName string) (int64, error)

	// GetGaugeValueByName - function for getting gauge metric value by it's name from PostgreSQL/in-memory storage.
	GetGaugeValueByName(metricName string) (float64, error)

	// CheckConnection - function for checking if repository is ok and available.
	CheckConnection(ctx context.Context) error

	// GetAllGaugeMetrics - function for getting all gauge metrics from PostgreSQL/in-memory storage.
	GetAllGaugeMetrics() (map[string]float64, error)

	// GetAllCounterMetrics - function for getting all counter metrics from PostgreSQL/in-memory storage.
	GetAllCounterMetrics() (map[string]int64, error)

	// RepositoryAddAllValues - function for updating all metrics as a batch of metrics in PostgreSQL/in-memory storage.
	RepositoryAddAllValues(metrics []data.Metrics) error

	CloseConnections() error
}
