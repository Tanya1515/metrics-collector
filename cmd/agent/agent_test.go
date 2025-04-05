package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMetricsUtil(t *testing.T) {
	
}

func TestGetMetrics(t *testing.T) {

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
