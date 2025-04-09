// Data contains basic types and their methods of the project.
package data

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

// ResultMetrics - type, that contains all gauge and counter metrics for representation them in HTML-format.
type ResultMetrics struct {
	// GaugeMetrics contains all gauge metrics.
	GaugeMetrics string
	// CounterMetrics contains all counter metrics.
	CounterMetrics string
}

// Metrics - type, that describes all fields of recieved/saved metrics.
type Metrics struct {
	// ID - name of the metric
	ID string `json:"id"`
	// MType - type of the metric, available values are counter and gauge
	MType string `json:"type"`
	// Delta - int64 value, that is specified for counter metric
	Delta *int64 `json:"delta,omitempty"`
	// Value - float64 value, that is specified for gauge metric
	Value *float64 `json:"value,omitempty"`
}

// Compress - function for compressing list of metrics to slice of bytes
func Compress(metricData *[]Metrics) ([]byte, error) {
	var b bytes.Buffer

	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed init compress writer: %w", err)
	}

	metricDataBytes, err := json.Marshal(metricData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	_, err = w.Write(metricDataBytes)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %w", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %w", err)
	}

	return b.Bytes(), nil
}
