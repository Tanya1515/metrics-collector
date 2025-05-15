package structure

import (
	"context"
	"sync"

	storage "github.com/Tanya1515/metrics-collector.git/cmd/storage"
	"github.com/pkg/errors"
)

var errorMetricExists = errors.New("ErrMetricExists")

// MemStorage - data structure for describing in-memory storage
type MemStorage struct {
	storage.StoreType
	counterStorage map[string]int64
	gaugeStorage   map[string]float64
	mutex          *sync.Mutex
}

func (S *MemStorage) Init(shutdown chan struct{}, Gctx context.Context) error {
	var mutex sync.Mutex
	S.counterStorage = make(map[string]int64, 1000)
	S.gaugeStorage = make(map[string]float64, 1000)
	S.mutex = &mutex

	if S.Restore {
		err := S.Store(S)
		if err != nil {
			return err
		}
	}

	if (S.FileStore != "") && (S.BackupTimer != 0) {

		go S.SaveMetricsAsync(shutdown, Gctx, S)
	}
	return nil
}

func (S *MemStorage) CheckConnection(ctx context.Context) error {
	return nil
}
