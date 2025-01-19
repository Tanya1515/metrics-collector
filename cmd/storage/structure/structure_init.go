package storage

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

var errorMetricExists = errors.New("ErrMetricExists")

type MemStorage struct {
	counterStorage map[string]int64
	gaugeStorage   map[string]float64
	mutex          *sync.Mutex
}

func (S *MemStorage) Init() error {
	var mutex sync.Mutex
	S.counterStorage = make(map[string]int64, 100)
	S.gaugeStorage = make(map[string]float64, 100)
	S.mutex = &mutex
	return nil
}

func (S *MemStorage) CheckConnection(ctx context.Context) error {
	return nil
}

func (S *MemStorage) CheckConnection(ctx context.Context) error {
	return nil
}
