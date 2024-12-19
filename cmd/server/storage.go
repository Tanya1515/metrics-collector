package main

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
