package storage

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var errorMetricExists = errors.New("ErrMetricExists")

type MemStorage struct {
	counterStorage map[string]int64
	gaugeStorage   map[string]float64
	mutex          *sync.Mutex
	backupTimer    time.Duration
	fileStore      string
}

func (S *MemStorage) Init(restore bool, fileStore string, backupTimer time.Duration) error {
	var mutex sync.Mutex
	S.counterStorage = make(map[string]int64, 100)
	S.gaugeStorage = make(map[string]float64, 100)
	S.fileStore = fileStore
	S.backupTimer = backupTimer
	S.mutex = &mutex

	if restore {
		err := S.Store()
		return err
	}
	if (S.fileStore != "") && (S.backupTimer != 0) {
		go S.SaveMetricsAsync()
	}
	return nil
}

func (S *MemStorage) CheckConnection(ctx context.Context) error {
	return nil
}
