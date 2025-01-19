package storage

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
