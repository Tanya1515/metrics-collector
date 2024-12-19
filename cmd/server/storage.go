package main

import "errors"

var errorMetricExists = errors.New("error: metric does not exist on the server side")

type MemStorage struct {
	CounterStorage map[string][]int64
	GaugeStorage   map[string]float64
}

type ReposutiryInterface interface {
	RepositoryAddCounterValue()
	RepositoryAddGaugeValue()
}

func (S *MemStorage) RepositoryAddCounterValue(metricName string, metricValue int64) {
	S.CounterStorage[metricName] = append(S.CounterStorage[metricName], metricValue)
}

func (S *MemStorage) RepositoryAddGaugeValue(metricName string, metricValue float64) {
	S.GaugeStorage[metricName] = metricValue
}

func (S *MemStorage) GetCounterValueByName(metricName string) ([]int64, error) {
	for key, value := range S.CounterStorage {
		if key == metricName {
			return value, nil
		}
	}
	return []int64{}, errorMetricExists
}

func (S *MemStorage) GetGaugeValueByName(metricName string) (float64, error) {
	for key, value := range S.GaugeStorage {
		if key == metricName {
			return value, nil
		}
	}
	return 0, errorMetricExists
}
