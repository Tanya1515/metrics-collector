package main

type MemStorage struct {
	CounterStorage map[string]int64
	GaugeStorage   map[string]float64
}

type MemStorageInterface interface {
	StorageAddCounterValue()
	StorageAddGaugeValue()
}

func (S *MemStorage) StorageAddCounterValue(metric_name string, metric_value int64) {
	S.CounterStorage[metric_name] = metric_value
}

func (S *MemStorage) StorageAddGaugeValue(metric_name string, metric_value float64) {
	S.GaugeStorage[metric_name] = metric_value
}