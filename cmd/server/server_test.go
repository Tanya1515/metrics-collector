package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"sync"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestProcessRequest(t *testing.T) {
	var mutex sync.Mutex
	var gaugeMetricValue = 1.5
	var counterMetrciValue int64 = 4
	type httpResult struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name    string
		request string
		metric  *Metrics
		storage *MemStorage
		result  httpResult
		modify  string
	}{
		{
			name:    "test: Send correct counter value",
			request: "/update",
			metric:  &Metrics{ID: "value", MType: "counter", Delta: &counterMetrciValue},
			storage: &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        200,
				response:    "{\"id\":\"value\",\"type\":\"counter\",\"delta\":4}",
				contentType: "application/json",
			},
			modify: "counter",
		},
		{
			name:    "test: Send correct gauge value",
			request: "/update",
			metric:  &Metrics{ID: "value", MType: "gauge", Value: &gaugeMetricValue},
			storage: &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        200,
				response:    "{\"id\":\"value\",\"type\":\"gauge\",\"value\":1.5}",
				contentType: "application/json",
			},
			modify: "gauge",
		},
		{
			name:    "test: Send incorrect metric type",
			request: "/update",
			metric:  &Metrics{ID: "value", MType: "test", Delta: &counterMetrciValue},
			storage: &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        400,
				response:    "Error 400: Invalid metric type: test\n",
				contentType: "text/plain; charset=utf-8",
			},
			modify: "",
		},
		{
			name:    "test: Send none metric name",
			request: "/update",
			metric:  &Metrics{ID: "", MType: "counter", Delta: &counterMetrciValue},
			storage: &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        404,
				response:    "Error 404: Metric name was not found\n",
				contentType: "text/plain; charset=utf-8",
			},
			modify: "",
		},
	}

	for _, test := range tests {
		var mutex sync.RWMutex
		t.Run("Test:", func(t *testing.T) {
			var buf bytes.Buffer
			bodyRequestEncode := json.NewEncoder(&buf)
			err := bodyRequestEncode.Encode(test.metric)
			if err != nil {
				panic(err)
			}
			request := httptest.NewRequest(http.MethodPost, test.request, &buf)

			logger, err := zap.NewDevelopment()
			if err != nil {
				panic(err)
			}

			defer logger.Sync()
			App := Application{Storage: test.storage, logger: *logger.Sugar()}

			w := httptest.NewRecorder()

			h := http.HandlerFunc(App.UpdateValue(&mutex))
			h(w, request)

			res := w.Result()
			assert.Equal(t, test.result.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, test.result.response, string(resBody))

			if test.modify == "counter" {
				assert.Contains(t, test.storage.CounterStorage, test.metric.ID)
			}

			if test.modify == "gauge" {
				assert.Contains(t, test.storage.GaugeStorage, test.metric.ID)
			}
			assert.Equal(t, test.result.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestGetMetric(t *testing.T) {
	var mutex sync.Mutex

	type httpResult struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name    string
		request string
		metric  *Metrics
		storage *MemStorage
		result  httpResult
	}{
		{
			name:    "test: Send incorrect metric type",
			request: "/value",
			metric:  &Metrics{ID: "PollCount", MType: "test"},
			storage: &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        400,
				response:    "Error 400: Invalid metric type: test\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "test: Get correct counter metric",
			request: "/value",
			metric:  &Metrics{ID: "PollCount", MType: "counter"},
			storage: &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        200,
				response:    "{\"id\":\"PollCount\",\"type\":\"counter\",\"delta\":1}",
				contentType: "application/json",
			},
		},
		{
			name:    "test: Get correct gauge metric",
			request: "/value",
			metric:  &Metrics{ID: "BuckHashSys", MType: "gauge"},
			storage: &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        200,
				response:    "{\"id\":\"BuckHashSys\",\"type\":\"gauge\",\"value\":0.1}",
				contentType: "application/json",
			},
		},
		{
			name:    "test: Get not existing gauge metric",
			request: "/value",
			metric:  &Metrics{ID: "GaugeTest", MType: "gauge"},
			storage: &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        404,
				response:    "Error 404: GaugeTest does not exist in gauge storage: ErrMetricExists\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "test: Get not existing counter metric",
			request: "/value",
			metric:  &Metrics{ID: "PollCountEx", MType: "counter"},
			storage: &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        404,
				response:    "Error 404: PollCountEx does not exist in counter storage: ErrMetricExists\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "test: Get request without metric name",
			request: "/value",
			metric:  &Metrics{ID: "", MType: "counter"},
			storage: &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        404,
				response:    "Error 404: Metric name was not found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, test := range tests {
		t.Run("Test:", func(t *testing.T) {
			var buf bytes.Buffer
			bodyRequestEncode := json.NewEncoder(&buf)
			err := bodyRequestEncode.Encode(test.metric)
			if err != nil {
				panic(err)
			}
			request := httptest.NewRequest(http.MethodGet, "/value", &buf)

			logger, err := zap.NewDevelopment()
			if err != nil {
				panic(err)
			}

			defer logger.Sync()
			App := Application{Storage: test.storage, logger: *logger.Sugar()}

			w := httptest.NewRecorder()

			h := http.HandlerFunc(App.GetMetric())
			h(w, request)

			res := w.Result()
			assert.Equal(t, test.result.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, test.result.response, string(resBody))
			assert.Equal(t, test.result.contentType, res.Header.Get("Content-Type"))
		})
	}
}
