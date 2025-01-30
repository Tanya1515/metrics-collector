package storage

import (
	"encoding/json"
	"os"
	"time"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
)

func (S *MemStorage) SaveMetricsAsync() {

	for {
		S.SaveMetrics()
		time.Sleep(S.backupTimer * time.Second)
	}
}

func (S *MemStorage) SaveMetrics() (err error) {
	allMetrics := make([]data.Metrics, len(S.counterStorage)+len(S.gaugeStorage))
	gaugeMetric := data.Metrics{ID: "", MType: "gauge"}
	counterMetric := data.Metrics{ID: "", MType: "counter"}
	i := 0
	allGaugeMetrics := S.GetAllGaugeMetrics()
	for metricName, metricValue := range allGaugeMetrics {
		gaugeMetric.ID = metricName
		gaugeMetric.Value = &metricValue
		allMetrics[i] = gaugeMetric
		i += 1
	}

	allCounterMetrics := S.GetAllCounterMetrics()
	for metricName, metricValue := range allCounterMetrics {
		counterMetric.ID = metricName
		counterMetric.Delta = &metricValue
		allMetrics[i] = counterMetric
		i += 1
	}

	metricsBytes, err := json.Marshal(allMetrics)
	if err != nil {
		return
	}
	err = os.WriteFile(S.fileStore, metricsBytes, 0644)
	if err != nil {
		return
	}

	return nil
}

func (S *MemStorage) Store() error {
	allMetrics := make([]data.Metrics, len(S.counterStorage)+len(S.gaugeStorage))

	dataFromFile, err := os.ReadFile(S.fileStore)
	if err != nil {
		return err
	}

	err = json.Unmarshal(dataFromFile, &allMetrics)
	if err != nil {
		return nil
	}

	for _, metric := range allMetrics {
		if metric.MType == "gauge" {
			S.RepositoryAddGaugeValue(metric.ID, *metric.Value)
		}

		if metric.MType == "counter" {
			S.RepositoryAddValue(metric.ID, *metric.Delta)
		}
	}

	return nil
}
