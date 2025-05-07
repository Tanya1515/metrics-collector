package storage

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

var errorMetricExists = errors.New("ErrMetricExists")

// MemStorage - data structure for describing in-memory storage 
type MemStorage struct {
	counterStorage map[string]int64
	gaugeStorage   map[string]float64
	mutex          *sync.Mutex
	backupTimer    int
	fileStore      string
}

func (S *MemStorage) Init(restore bool, fileStore string, backupTimer int) error {
	var mutex sync.Mutex
	S.counterStorage = make(map[string]int64, 1000)
	S.gaugeStorage = make(map[string]float64, 1000)
	S.fileStore = fileStore
	S.backupTimer = backupTimer
	S.mutex = &mutex

	if restore {
		err := S.Store()
		if err != nil {
			return err
		}
	}

	if (S.fileStore != "") && (S.backupTimer != 0) {

		go S.SaveMetricsAsync()
	}
	return nil
}

func (S *MemStorage) CheckConnection(ctx context.Context) error {
	return nil
}
