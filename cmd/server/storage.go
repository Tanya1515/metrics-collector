package main

import (
	"sync"

	"github.com/pkg/errors"
)

var errorMetricExists = errors.New("ErrMetricExists")

type MemStorage struct {
	counterStorage map[string]int64
	gaugeStorage   map[string]float64
	mutex          *sync.Mutex
	backup         bool
	fileStore      string
}

type ReposutiryInterface interface {
	RepositoryAddCounterValue()
	RepositoryAddGaugeValue()
}

func (S *MemStorage) RepositoryAddValue(metricName string, metricValue int64) {
	S.mutex.Lock()

	defer S.mutex.Unlock()
	S.counterStorage[metricName] = metricValue
}

func (S *MemStorage) RepositoryAddCounterValue(metricName string, metricValue int64) {
	S.mutex.Lock()

	defer S.mutex.Unlock()
	S.counterStorage[metricName] = S.counterStorage[metricName] + metricValue
}

func (S *MemStorage) RepositoryAddGaugeValue(metricName string, metricValue float64) {
	S.mutex.Lock()

	defer S.mutex.Unlock()
	S.gaugeStorage[metricName] = metricValue
}

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

func (S *MemStorage) GetAllGaugeMetrics() map[string]float64 {
	S.mutex.Lock()

	defer S.mutex.Unlock()
	AllGaugeMetrics := S.gaugeStorage
	return AllGaugeMetrics
}

func (S *MemStorage) GetAllCounterMetrics() map[string]int64 {
	S.mutex.Lock()

	defer S.mutex.Unlock()

	AllCounterMetrics := S.counterStorage
	return AllCounterMetrics
}
