package storage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
)

type StoreStorage interface {
	SaveMetricsAsync(Gctx context.Context)

	SaveMetrics() (err error)

	Store() error
}

type StoreType struct {
	Restore     bool
	BackupTimer int
	FileStore   string
	Shutdown    chan struct{}
}

// SaveMetricsAsync - function for saving metrics every
func (S *StoreType) SaveMetricsAsync(Gctx context.Context, storage RepositoryInterface) {
	for {
		select {
		case <-Gctx.Done():
			close(S.Shutdown)
			return
		default:
			S.SaveMetrics(storage)
			time.Sleep(time.Duration(S.BackupTimer) * time.Second)
		}
	}
}

// SaveMetrics - function for saving metrics into file asynchronously.
func (S *StoreType) SaveMetrics(storage RepositoryInterface) (err error) {
	allMetrics := make([]data.Metrics, 100)
	gaugeMetric := data.Metrics{ID: "", MType: "gauge"}
	counterMetric := data.Metrics{ID: "", MType: "counter"}
	i := 0
	allGaugeMetrics, err := storage.GetAllGaugeMetrics()
	if err != nil {
		return
	}
	for metricName, metricValue := range allGaugeMetrics {
		gaugeMetric.ID = metricName
		gaugeMetric.Value = &metricValue
		if i < len(allMetrics) {
			allMetrics[i] = gaugeMetric
			i += 1
		} else {
			allMetrics = append(allMetrics, gaugeMetric)
		}
	}

	allCounterMetrics, err := storage.GetAllCounterMetrics()
	if err != nil {
		return
	}
	for metricName, metricValue := range allCounterMetrics {

		counterMetric.ID = metricName
		counterMetric.Delta = &metricValue
		if i < len(allMetrics) {
			allMetrics[i] = counterMetric
			i += 1
		} else {
			allMetrics = append(allMetrics, counterMetric)
		}

	}

	metricsBytes, err := json.Marshal(allMetrics)
	if err != nil {
		return
	}

	err = os.WriteFile(S.FileStore, metricsBytes, 0644)
	if err != nil {
		return
	}

	return nil
}

// Store - function for initialization in-memory storage from backup file.
func (S *StoreType) Store(storage RepositoryInterface) error {
	allMetrics := make([]data.Metrics, 100)

	_, err := os.Stat(S.FileStore)
	if errors.Is(err, os.ErrNotExist) {
		_, err = os.Create(S.FileStore)
		if err != nil {
			return err
		}
	}

	dataFromFile, err := os.ReadFile(S.FileStore)
	if err != nil {
		return err
	}

	err = json.Unmarshal(dataFromFile, &allMetrics)
	if err != nil {
		return nil
	}

	for _, metric := range allMetrics {
		if metric.MType == "gauge" {
			storage.RepositoryAddGaugeValue(metric.ID, *metric.Value)
		}

		if metric.MType == "counter" {
			storage.RepositoryAddValue(metric.ID, *metric.Delta)
		}
	}

	return nil
}
