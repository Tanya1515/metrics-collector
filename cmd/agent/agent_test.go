package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckValue(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		result    bool
	}{
		{
			name:      "test: Name of Metric exists",
			fieldName: "Alloc",
			result:    true,
		},
		{
			name:      "test: Name of Metric does not exist",
			fieldName: "Test",
			result:    false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := CheckValue(test.fieldName)
			assert.Equal(t, test.result, result)
		})
	}
}

// func TestGetMetrics(t *testing.T) {
// 	type testResult struct {
// 		mapMetricslen int
// 		PollCount     int
// 	}
// 	tests := []struct {
// 		name       string
// 		mapMetrics map[string]interface{}
// 		PollCount  int
// 		timer      time.Duration
// 		result     testResult
// 	}{
// 		{
// 			name:       "test: Get Metrics",
// 			mapMetrics: make(map[string]interface{}, 20),
// 			PollCount:  0,
// 			timer:      2,
// 			result: testResult{
// 				mapMetricslen: 28,
// 				PollCount:     1,
// 			},
// 		},
// 	}

// 	var mutex sync.RWMutex
// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			GetMetrics(&test.mapMetrics, &test.PollCount, test.timer, &mutex)
// 			assert.Equal(t, test.result.mapMetricslen, len(test.mapMetrics))
// 			assert.Equal(t, test.result.PollCount, test.PollCount)
// 		})
// 	}
// }

func TestMakeString(t *testing.T) {
	tests := []struct {
		name          string
		metricName    string
		metricValue   string
		metricType    string
		serverAddress string
		result        string
	}{
		{
			name:          "test: Make http-request with gauge metric",
			metricName:    "Alloc",
			metricValue:   "1.5",
			metricType:    "gauge",
			serverAddress: "localhost:8080",
			result:        "http://localhost:8080/update/gauge/Alloc/1.5",
		},
		{
			name:          "test: Make http-request with counter metric",
			metricName:    "Alloc",
			metricValue:   "10",
			metricType:    "counter",
			serverAddress: "192.168.0.1:8085",
			result:        "http://192.168.0.1:8085/update/counter/PollCountAlloc/10",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := MakeString(test.serverAddress, test.metricName, test.metricValue, test.metricType)
			assert.Equal(t, test.result, result)
		})
	}
}
