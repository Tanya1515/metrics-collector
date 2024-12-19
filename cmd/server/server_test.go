package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

		
func TestProcessRequest(t *testing.T) {
	type http_result struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name       string
		metricInfo string
		storage    *MemStorage
		result     http_result
		modify     string
	}{
		{
			name: "test: Send correct counter value",
			metricInfo: "/update/counter/value/4",
			storage: &MemStorage{CounterStorage: map[string][]int64{"PollCount": {1, 2, 3}}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}},
			result: http_result{
				code:        200,
				response:    "Succesfully edit!",
				contentType: "text/plain; charset=utf-8",
			},
			modify: "counter",
		},
		{
			name: "test: Send incorrect metric type",
			metricInfo: "/update/test/value/1.5",
			storage: &MemStorage{CounterStorage: map[string][]int64{"PollCount": {1,2,3}}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}},
			result: http_result{
				code: 400,
				response: "Error 400: Invalid metric type: test\n",
				contentType: "text/plain; charset=utf-8",
			},
			modify: "",
		},
		{
			name: "test: Send correct gauge value",
			metricInfo: "/update/gauge/value/1.5",
			storage: &MemStorage{CounterStorage: map[string][]int64{"PollCount": {1,2,3}}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}},
			result: http_result{
				code: 200,
				response: "Succesfully edit!",
				contentType: "text/plain; charset=utf-8",
			},
			modify: "gauge",
		},
		{
			name: "test: Send incorrect counter value",
			metricInfo: "/update/counter/value/1.5",
			storage: &MemStorage{CounterStorage: map[string][]int64{"PollCount": {1,2,3}}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}},
			result: http_result{
					code: 400,
					response: "Error 400: Invalid metric value: 1.5\n",
					contentType: "text/plain; charset=utf-8",
				},
			modify: "",
		},

		{
			name: "test: Send none metric name",
			metricInfo: "/update/counter//1",
			storage: &MemStorage{CounterStorage: map[string][]int64{"PollCount": {1,2,3}}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}},
			result: http_result{
				code: 404,
				response: "Error 404: Metric name was not found\n",
				contentType: "text/plain; charset=utf-8",
			},
			modify: "",
		},
	}

	for _, test := range tests {
		t.Run("Test 1:", func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.metricInfo, nil)

			w := httptest.NewRecorder()

			h := http.HandlerFunc(ProcessRequest(test.storage))
			h(w, request)

			res := w.Result()
			assert.Equal(t, test.result.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, test.result.response, string(resBody))

			if test.modify == "counter"{
				assert.Contains(t, test.storage.CounterStorage, (strings.Split(test.metricInfo, "/"))[3])
			}

			if test.modify == "gauge"{
				assert.Contains(t, test.storage.GaugeStorage, (strings.Split(test.metricInfo, "/"))[3])
			}
			assert.Equal(t, test.result.contentType, res.Header.Get("Content-Type"))
		})
	}
}
