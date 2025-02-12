package storage

import (
	"github.com/pkg/errors"
)

func (S *MemStorage) GetCounterValueByName(metricName string) (int64, error) {
	S.mutex.Lock()

	defer S.mutex.Unlock()
	for key, value := range S.counterStorage {
		if key == metricName {
			return value, nil
		}
	}
	return 0, errors.Wrapf(errorMetricExists, "%s does not exist in counter storage", metricName)
}

func (S *MemStorage) GetGaugeValueByName(metricName string) (float64, error) {
	S.mutex.Lock()

	defer S.mutex.Unlock()
	for key, value := range S.gaugeStorage {
		if key == metricName {
			return value, nil
		}
	}
	return 0, errors.Wrapf(errorMetricExists, "%s does not exist in gauge storage", metricName)
}

func (S *MemStorage) GetAllGaugeMetrics() (map[string]float64, error) {
	S.mutex.Lock()

	defer S.mutex.Unlock()

	AllGaugeMetrics := make(map[string]float64, len(S.gaugeStorage))

	for valueName, value := range S.gaugeStorage {
		AllGaugeMetrics[valueName] = value
	}
	return AllGaugeMetrics, nil
}

func (S *MemStorage) GetAllCounterMetrics() (map[string]int64, error) {
	S.mutex.Lock()

	defer S.mutex.Unlock()

	AllCounterMetrics := make(map[string]int64, len(S.counterStorage))
	for valueName, value := range S.counterStorage {
		AllCounterMetrics[valueName] = value
	}

	return AllCounterMetrics, nil
}
