package data

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
)

type ResultMetrics struct {
	GaugeMetrics   string
	CounterMetrics string
}

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (metricData *Metrics) Compress() ([]byte, error) {
	var b bytes.Buffer

	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed init compress writer: %v", err)
	}

	metricDataBytes, err := json.Marshal(metricData)
	if err != nil {
		return nil, fmt.Errorf("failed to compress data: %v", err)
	}
	_, err = w.Write(metricDataBytes)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}

	return b.Bytes(), nil
}
