package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckValue(t * testing.T){
	tests := []struct {
		name  string
		fieldName string
		result bool
	}{
		{
			name: "test: Name of Metric exists",
			fieldName: "Alloc",
			result: true,
		}, 
		{
			name: "test: Name of Metric does not exist",
			fieldName: "Test", 
			result: false,
		},
	}
	for _, test := range tests{
		t.Run(test.name, func(t *testing.T){
			result := CheckValue(test.fieldName)
			assert.Equal(t, test.result, result)
		})
	}
}

func TestGetMetrics(t * testing.T){
	type test_result struct {
		mapMetricslen int
		PollCount int
	}
	tests := []struct{
		name string 
		mapMetrics map[string]interface{}
		PollCount int
		result test_result
	}{
		{
			name: "test: Get Metrics",
			mapMetrics: make(map[string]interface{}, 20),
			PollCount: 0,
			result: test_result{
				mapMetricslen: 28,
				PollCount: 1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T){
			GetMetrics(&test.mapMetrics, &test.PollCount)
			assert.Equal(t, test.result.mapMetricslen, len(test.mapMetrics))
			assert.Equal(t, test.result.PollCount, test.PollCount)
		})
	}
}

func TestMakeString(t *testing.T){
	tests := []struct{
		name string
		metricName string
		metricValue string 
		metricType string
		result string
	}{
		{
			name: "test: Make http-request with gauge metric",
			metricName: "Alloc",
			metricValue: "1.5",
			metricType: "gauge",
			result: "http://127.0.0.1:8080/update/gauge/Alloc/1.5",
		},
		{
			name: "test: Make http-request with counter metric",
			metricName: "Alloc",
			metricValue: "10",
			metricType: "counter",
			result: "http://127.0.0.1:8080/update/counter/PollCountAlloc/10",
		},
	}

	for _, test := range tests{
		t.Run(test.name, func(t *testing.T){
			result := MakeString(test.metricName, test.metricValue, test.metricType)
			assert.Equal(t, test.result, result)
		})
	}
}