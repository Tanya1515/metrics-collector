package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessRequest(t *testing.T) {
	var mutex sync.Mutex

	type httpResult struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name        string
		metricInfo  string
		metricType  string
		metricName  string
		metricValue string
		storage     *MemStorage
		result      httpResult
		modify      string
	}{
		{
			name:        "test: Send correct counter value",
			metricInfo:  "/update",
			metricType:  "counter",
			metricName:  "value",
			metricValue: "4",
			storage:     &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        200,
				response:    "Succesfully edit!",
				contentType: "text/plain; charset=utf-8",
			},
			modify: "counter",
		},
		{
			name:        "test: Send incorrect metric type",
			metricInfo:  "/update",
			metricType:  "test",
			metricName:  "value",
			metricValue: "1.5",
			storage:     &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        400,
				response:    "Error 400: Invalid metric type: test\n",
				contentType: "text/plain; charset=utf-8",
			},
			modify: "",
		},
		{
			name:        "test: Send correct gauge value",
			metricInfo:  "/update",
			metricType:  "gauge",
			metricName:  "value",
			metricValue: "1.5",
			storage:     &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        200,
				response:    "Succesfully edit!",
				contentType: "text/plain; charset=utf-8",
			},
			modify: "gauge",
		},
		{
			name:        "test: Send incorrect counter value",
			metricInfo:  "/update",
			metricType:  "counter",
			metricName:  "value",
			metricValue: "1.5",
			storage:     &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        400,
				response:    "Error 400: Invalid metric value: 1.5\n",
				contentType: "text/plain; charset=utf-8",
			},
			modify: "",
		},

		{
			name:        "test: Send none metric name",
			metricInfo:  "/update",
			metricType:  "counter",
			metricName:  "",
			metricValue: "1",
			storage:     &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        404,
				response:    "Error 404: Metric name was not found\n",
				contentType: "text/plain; charset=utf-8",
			},
			modify: "",
		},
	}

	for _, test := range tests {
		t.Run("Test:", func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, test.metricInfo, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("metricType", test.metricType)
			rctx.URLParams.Add("metricName", test.metricName)
			rctx.URLParams.Add("metricValue", test.metricValue)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			App := Application{Storage: test.storage}

			w := httptest.NewRecorder()

			h := http.HandlerFunc(App.ProcessRequest())
			h(w, request)

			res := w.Result()
			assert.Equal(t, test.result.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, test.result.response, string(resBody))

			if test.modify == "counter" {
				assert.Contains(t, test.storage.CounterStorage, test.metricName)
			}

			if test.modify == "gauge" {
				assert.Contains(t, test.storage.GaugeStorage, test.metricName)
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
		name       string
		metricInfo string
		metricType string
		metricName string
		storage    *MemStorage
		result     httpResult
	}{
		{
			name:       "test: Send incorrect metric type",
			metricInfo: "/value",
			metricType: "test",
			metricName: "PollCount",
			storage:    &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        400,
				response:    "Error 400: Invalid metric type: test\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "test: Get correct counter metric",
			metricInfo: "/value",
			metricType: "counter",
			metricName: "PollCount",
			storage:    &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        200,
				response:    "1",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "test: Get correct gauge metric",
			metricInfo: "/value",
			metricType: "gauge",
			metricName: "BuckHashSys",
			storage:    &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        200,
				response:    "0.1",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "test: Get not existing gauge metric",
			metricInfo: "/value",
			metricType: "gauge",
			metricName: "GaugeTest",
			storage:    &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        404,
				response:    "Error 404: GaugeTest does not exist in gauge storage: ErrMetricExists\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "test: Get not existing counter metric",
			metricInfo: "/value",
			metricType: "counter",
			metricName: "PollCountEx",
			storage:    &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        404,
				response:    "Error 404: PollCountEx does not exist in counter storage: ErrMetricExists\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "test: Get request without metric name",
			metricInfo: "/value",
			metricType: "counter",
			metricName: "",
			storage:    &MemStorage{CounterStorage: map[string]int64{"PollCount": 1}, GaugeStorage: map[string]float64{"BuckHashSys": 0.1}, mutex: &mutex},
			result: httpResult{
				code:        404,
				response:    "Error 404: Metric name was not found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, test := range tests {
		t.Run("Test:", func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/value", nil)
			// add path parameters to chi request: https://github.com/go-chi/chi/blob/91a3777c41c3d3493a446f690b572d93a76cba73/mux_test.go#L1143-L1145
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("metricType", test.metricType)
			rctx.URLParams.Add("metricName", test.metricName)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			App := Application{Storage: test.storage}

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
