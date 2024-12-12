package main

type MemStorage struct {
	CounterStorage map[string]int64
	GaugeStorage   map[string]float64
}

type MemStorageInterface interface {
	StorageAddCounterValue()
	StorageAddGaugeValue()
}

func (S *MemStorage) StorageAddCounterValue(metricName string, metricValue int64) {
	S.CounterStorage[metricName] = metricValue
}

func (S *MemStorage) StorageAddGaugeValue(metricName string, metricValue float64) {
	S.GaugeStorage[metricName] = metricValue
}