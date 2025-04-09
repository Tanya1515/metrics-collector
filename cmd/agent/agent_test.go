package agent

import (
	"testing"
	"time"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
	"github.com/stretchr/testify/assert"
)

func TestGetMetricsUtil(t *testing.T) {
	chanSend := make(chan int64)
	chanMetrics := make(chan []data.Metrics)
	result := make([]string, 3)

	test := struct {
		name  string
		timer time.Duration
	}{
		name:  "test: collect metrics",
		timer: 1,
	}

	t.Run(test.name, func(t *testing.T) {
		go GetMetricsUtil(chanSend, chanMetrics, test.timer)
		chanSend <- -1
		metrics := <-chanMetrics

		for _, metric := range metrics {
			result = append(result, metric.ID)
		}
		assert.Contains(t, result, "TotalMemory")
		assert.Contains(t, result, "FreeMemory")
		assert.Contains(t, result, "CPUutilization1")
	})
}

func TestGetMetrics(t *testing.T) {
	chanSend := make(chan int64)
	chanMetrics := make(chan []data.Metrics)
	result := make([]string, 3)

	test := struct {
		name  string
		timer time.Duration
	}{
		name:  "test: collect metrics",
		timer: 1,
	}

	t.Run(test.name, func(t *testing.T) {
		go GetMetrics(chanSend, chanMetrics, test.timer)
		chanSend <- -1
		metrics := <-chanMetrics

		for _, metric := range metrics {
			result = append(result, metric.ID)
		}
		assert.Contains(t, result, "TotalAlloc")
		assert.Contains(t, result, "RandomValue")
		assert.Contains(t, result, "HeapReleased")
		assert.Contains(t, result, "Lookups")
		assert.Contains(t, result, "MCacheInuse")
	})
}

func TestMakeMetrics(t *testing.T) {
	tests := []struct {
		name       string
		mapMetrics map[string]float64
		pollCount  int64
	}{
		{
			name:       "test: make map with pollCount 10",
			mapMetrics: map[string]float64{"Alloc": 10.5, "BuckHashSys": 11, "Frees": 0.3, "GCCPUFraction": 5.7},
			pollCount:  10,
		},
		{
			name:       "test: make map with pollCount 125",
			mapMetrics: map[string]float64{"HeapIdle": 1.5, "HeapSys": 0.11, "MCacheSys": 0.4, "NumGC": 5.7},
			pollCount:  125,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := MakeMetrics(test.mapMetrics, test.pollCount)
			assert.Equal(t, *result[4].Delta, test.pollCount)
		})
	}

}

func TestMakeString(t *testing.T) {
	tests := []struct {
		name          string
		serverAddress string
		result        string
	}{
		{
			name:          "test: Make http-request with gauge metric",
			serverAddress: "localhost:8080",
			result:        "http://localhost:8080/updates/",
		},
		{
			name:          "test: Make http-request with counter metric",
			serverAddress: "192.168.0.1:8085",
			result:        "http://192.168.0.1:8085/updates/",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := MakeString(test.serverAddress)
			assert.Equal(t, test.result, result)
		})
	}
}
