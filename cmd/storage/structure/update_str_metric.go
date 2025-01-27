package storage

import (
	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
)

func (S *MemStorage) RepositoryAddValue(metricName string, metricValue int64) error {
	S.mutex.Lock()

	defer S.mutex.Unlock()
	S.counterStorage[metricName] = metricValue

	return nil
}

func (S *MemStorage) RepositoryAddCounterValue(metricName string, metricValue int64) error {
	S.mutex.Lock()

	defer S.mutex.Unlock()
	S.counterStorage[metricName] = S.counterStorage[metricName] + metricValue

	return nil
}

func (S *MemStorage) RepositoryAddGaugeValue(metricName string, metricValue float64) error {
	S.mutex.Lock()

	defer S.mutex.Unlock()
	S.gaugeStorage[metricName] = metricValue

	return nil
}

func (S *MemStorage) RepositoryAddAllValues(metrics []data.Metrics) error {

	for _, metric := range metrics{
		if metric.MType == "counter"{
			S.RepositoryAddCounterValue(metric.ID, *metric.Delta)
		} else if metric.MType == "gauge" {
			S.RepositoryAddGaugeValue(metric.ID, *metric.Value)
		}
	}
	return nil
}
