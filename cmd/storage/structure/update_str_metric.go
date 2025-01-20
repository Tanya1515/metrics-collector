package storage

func (S *MemStorage) RepositoryAddValue(metricName string, metricValue int64) error {
	S.mutex.Lock()
	S.counterStorage[metricName] = metricValue
	S.mutex.Unlock()

	if (S.fileStore != "") && (S.backupTimer == 0) {
		S.SaveMetrics()
	}

	return nil
}

func (S *MemStorage) RepositoryAddCounterValue(metricName string, metricValue int64) error {
	S.mutex.Lock()
	S.counterStorage[metricName] = S.counterStorage[metricName] + metricValue
	S.mutex.Unlock()

	if (S.fileStore != "") && (S.backupTimer == 0) {
		S.SaveMetrics()
	}

	return nil
}

func (S *MemStorage) RepositoryAddGaugeValue(metricName string, metricValue float64) error {
	S.mutex.Lock()
	S.gaugeStorage[metricName] = metricValue
	S.mutex.Unlock()

	if (S.fileStore != "") && (S.backupTimer == 0) {
		S.SaveMetrics()
	}

	return nil
}
