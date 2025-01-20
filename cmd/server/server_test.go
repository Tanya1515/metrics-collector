package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	str "github.com/Tanya1515/metrics-collector.git/cmd/storage/structure"
)

func TestProcessRequest(t *testing.T) {
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
		storage *str.MemStorage
		result  httpResult
		modify  string
	}{
		{
			name:    "test: Send correct counter value",
			request: "/update",
			metric:  &Metrics{ID: "value", MType: "counter", Delta: &counterMetrciValue},
			storage: &str.MemStorage{},
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
			storage: &str.MemStorage{},
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
			storage: &str.MemStorage{},
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
			storage: &str.MemStorage{},
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
			err := test.storage.Init()
			if err != nil {
				panic(err)
			}
			test.storage.RepositoryAddCounterValue("PollCount", 1)
			test.storage.RepositoryAddGaugeValue("BuckHashSys", 0.1)
			var buf bytes.Buffer
			bodyRequestEncode := json.NewEncoder(&buf)
			err = bodyRequestEncode.Encode(test.metric)
			if err != nil {
				panic(err)
			}
			request := httptest.NewRequest(http.MethodPost, test.request, &buf)

			logger, err := zap.NewDevelopment()
			if err != nil {
				panic(err)
			}

			defer logger.Sync()
			App := Application{Storage: test.storage, Logger: *logger.Sugar()}

			w := httptest.NewRecorder()

			h := http.HandlerFunc(App.UpdateValue())
			h(w, request)

			res := w.Result()
			assert.Equal(t, test.result.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, test.result.response, string(resBody))

			if test.modify == "counter" {
				value, _ := test.storage.GetCounterValueByName(test.metric.ID)
				assert.Equal(t, value, *test.metric.Delta)
			}

			if test.modify == "gauge" {
				value, _ := test.storage.GetGaugeValueByName(test.metric.ID)
				assert.Equal(t, value, *test.metric.Value)
			}
			assert.Equal(t, test.result.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestGetMetric(t *testing.T) {
	type httpResult struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name    string
		request string
		metric  *Metrics
		storage *str.MemStorage
		result  httpResult
	}{
		{
			name:    "test: Send incorrect metric type",
			request: "/value",
			metric:  &Metrics{ID: "PollCount", MType: "test"},
			storage: &str.MemStorage{},
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
			storage: &str.MemStorage{},
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
			storage: &str.MemStorage{},
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
			storage: &str.MemStorage{},
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
			storage: &str.MemStorage{},
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
			storage: &str.MemStorage{},
			result: httpResult{
				code:        404,
				response:    "Error 404: Metric name was not found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, test := range tests {
		t.Run("Test:", func(t *testing.T) {
			err := test.storage.Init()
			if err != nil {
				panic(err)
			}
			test.storage.RepositoryAddCounterValue("PollCount", 1)
			test.storage.RepositoryAddGaugeValue("BuckHashSys", 0.1)
			var buf bytes.Buffer
			bodyRequestEncode := json.NewEncoder(&buf)
			err = bodyRequestEncode.Encode(test.metric)
			if err != nil {
				panic(err)
			}
			request := httptest.NewRequest(http.MethodGet, "/value", &buf)

			logger, err := zap.NewDevelopment()
			if err != nil {
				panic(err)
			}

			defer logger.Sync()
			App := Application{Storage: test.storage, Logger: *logger.Sugar()}

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