package main

import (
	"sync"

	"github.com/pkg/errors"
)

var errorMetricExists = errors.New("ErrMetricExists")

type MemStorage struct {
	CounterStorage map[string]int64
	GaugeStorage   map[string]float64
	mutex          *sync.Mutex
}

type ReposutiryInterface interface {
	RepositoryAddCounterValue()
	RepositoryAddGaugeValue()
}

func (S *MemStorage) RepositoryAddCounterValue(metricName string, metricValue int64) {
	S.CounterStorage[metricName] = S.CounterStorage[metricName] + metricValue
}

func (S *MemStorage) RepositoryAddGaugeValue(metricName string, metricValue float64) {
	S.GaugeStorage[metricName] = metricValue
}

func (S *MemStorage) GetCounterValueByName(metricName string) (int64, error) {
	for key, value := range S.CounterStorage {
		if key == metricName {
			return value, nil
		}
	}
	return 0, errors.Wrapf(errorMetricExists, "%s does not exist in counter storage", metricName)
}

func (S *MemStorage) GetGaugeValueByName(metricName string) (float64, error) {
	for key, value := range S.GaugeStorage {
		if key == metricName {
			return value, nil
		}
	}
	return 0, errors.Wrapf(errorMetricExists, "%s does not exist in gauge storage", metricName)
}
