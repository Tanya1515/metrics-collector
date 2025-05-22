package structure

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	storage "github.com/Tanya1515/metrics-collector.git/cmd/storage"
)

var errorMetricExists = errors.New("ErrMetricExists")

// MemStorage - data structure for describing in-memory storage
type MemStorage struct {
	storage.StoreType
	counterStorage map[string]int64
	gaugeStorage   map[string]float64
	mutex          *sync.Mutex
}

func (S *MemStorage) Init(Gctx context.Context, shutdown chan struct{}) error {
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

		go S.SaveMetricsAsync(Gctx, S)
	}
	return nil
}

func (S *MemStorage) CheckConnection(ctx context.Context) error {
	return nil
}

func (S *MemStorage) CloseConnections() error {
	<-S.Shutdown
	return nil
}
