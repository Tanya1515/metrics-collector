package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
	str "github.com/Tanya1515/metrics-collector.git/cmd/storage/structure"
)

func TestUpdateValuePath(t *testing.T) {

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
		storage     *str.MemStorage
		result      httpResult
		modify      string
	}{
		{
			name:        "test: Send correct counter value",
			metricInfo:  "/update",
			metricType:  "counter",
			metricName:  "value",
			metricValue: "4",
			storage:     &str.MemStorage{},
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
			storage:     &str.MemStorage{},
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
			storage:     &str.MemStorage{},
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
			storage:     &str.MemStorage{},
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
			storage:     &str.MemStorage{},
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
			chanSh := make(chan struct{})
			err := test.storage.Init(false, "", 0, chanSh, context.Background())
			if err != nil {
				panic(err)
			}
			test.storage.RepositoryAddCounterValue("PollCount", 1)
			test.storage.RepositoryAddGaugeValue("BuckHashSys", 0.1)
			request := httptest.NewRequest(http.MethodPost, test.metricInfo, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("metricType", test.metricType)
			rctx.URLParams.Add("metricName", test.metricName)
			rctx.URLParams.Add("metricValue", test.metricValue)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			logger, err := zap.NewDevelopment()
			if err != nil {
				panic(err)
			}

			defer logger.Sync()
			App := Application{Storage: test.storage, Logger: *logger.Sugar()}

			w := httptest.NewRecorder()

			h := http.HandlerFunc(App.UpdateValuePath())
			h(w, request)

			res := w.Result()
			assert.Equal(t, test.result.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, test.result.response, string(resBody))

			if test.modify == "counter" {
				value, _ := test.storage.GetCounterValueByName(test.metricName)
				metricValueInt64, _ := strconv.ParseInt(test.metricValue, 10, 64)
				assert.Equal(t, value, metricValueInt64)
			}

			if test.modify == "gauge" {
				value, _ := test.storage.GetGaugeValueByName(test.metricName)
				metricValueFloat64, _ := strconv.ParseFloat(test.metricValue, 64)
				assert.Equal(t, value, metricValueFloat64)
			}

			assert.Equal(t, test.result.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestUpdateValue(t *testing.T) {
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
		metric  *data.Metrics
		storage *str.MemStorage
		result  httpResult
		modify  string
	}{
		{
			name:    "test: Send correct counter value",
			request: "/update/",
			metric:  &data.Metrics{ID: "value", MType: "counter", Delta: &counterMetrciValue},
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
			request: "/update/",
			metric:  &data.Metrics{ID: "value", MType: "gauge", Value: &gaugeMetricValue},
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
			request: "/update/",
			metric:  &data.Metrics{ID: "value", MType: "test", Delta: &counterMetrciValue},
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
			request: "/update/",
			metric:  &data.Metrics{ID: "", MType: "counter", Delta: &counterMetrciValue},
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
			chanSh := make(chan struct{})
			err := test.storage.Init(false, "", 0, chanSh, context.Background())
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
		metric  *data.Metrics
		storage *str.MemStorage
		result  httpResult
	}{
		{
			name:    "test: Send incorrect metric type",
			request: "/value/",
			metric:  &data.Metrics{ID: "PollCount", MType: "test"},
			storage: &str.MemStorage{},
			result: httpResult{
				code:        400,
				response:    "Error 400: Invalid metric type: test\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "test: Get correct counter metric",
			request: "/value/",
			metric:  &data.Metrics{ID: "PollCount", MType: "counter"},
			storage: &str.MemStorage{},
			result: httpResult{
				code:        200,
				response:    "{\"id\":\"PollCount\",\"type\":\"counter\",\"delta\":1}",
				contentType: "application/json",
			},
		},
		{
			name:    "test: Get correct gauge metric",
			request: "/value/",
			metric:  &data.Metrics{ID: "BuckHashSys", MType: "gauge"},
			storage: &str.MemStorage{},
			result: httpResult{
				code:        200,
				response:    "{\"id\":\"BuckHashSys\",\"type\":\"gauge\",\"value\":0.1}",
				contentType: "application/json",
			},
		},
		{
			name:    "test: Get not existing gauge metric",
			request: "/value/",
			metric:  &data.Metrics{ID: "GaugeTest", MType: "gauge"},
			storage: &str.MemStorage{},
			result: httpResult{
				code:        404,
				response:    "Error 404: GaugeTest does not exist in gauge storage: ErrMetricExists\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "test: Get not existing counter metric",
			request: "/value/",
			metric:  &data.Metrics{ID: "PollCountEx", MType: "counter"},
			storage: &str.MemStorage{},
			result: httpResult{
				code:        404,
				response:    "Error 404: PollCountEx does not exist in counter storage: ErrMetricExists\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "test: Get request without metric name",
			request: "/value/",
			metric:  &data.Metrics{ID: "", MType: "counter"},
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
			chanSh := make(chan struct{})
			err := test.storage.Init(false, "", 0, chanSh, context.Background())
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
			request := httptest.NewRequest(http.MethodPost, "/value", &buf)

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

func TestGetMetricPath(t *testing.T) {
	type httpResult struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name       string
		request    string
		metricType string
		metricName string
		storage    *str.MemStorage
		result     httpResult
	}{
		{
			name:       "test: Send incorrect metric type",
			request:    "/value/",
			metricType: "test",
			metricName: "PollCount",
			storage:    &str.MemStorage{},
			result: httpResult{
				code:        400,
				response:    "Error 400: Invalid metric type: test\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "test: Get correct counter metric",
			request:    "/value/",
			metricType: "counter",
			metricName: "PollCount",
			storage:    &str.MemStorage{},
			result: httpResult{
				code:        200,
				response:    "1",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "test: Get correct gauge metric",
			request:    "/value/",
			metricType: "gauge",
			metricName: "BuckHashSys",
			storage:    &str.MemStorage{},
			result: httpResult{
				code:        200,
				response:    "0.1",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "test: Get not existing gauge metric",
			request:    "/value/",
			metricType: "gauge",
			metricName: "GaugeTest",
			storage:    &str.MemStorage{},
			result: httpResult{
				code:        404,
				response:    "Error 404: GaugeTest does not exist in gauge storage: ErrMetricExists\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "test: Get not existing counter metric",
			request:    "/value/",
			metricType: "counter",
			metricName: "PollCountEx",
			storage:    &str.MemStorage{},
			result: httpResult{
				code:        404,
				response:    "Error 404: PollCountEx does not exist in counter storage: ErrMetricExists\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "test: Get request without metric name",
			request:    "/value/",
			metricType: "counter",
			metricName: "",
			storage:    &str.MemStorage{},
			result: httpResult{
				code:        404,
				response:    "Error 404: Metric name was not found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, test := range tests {
		t.Run("Test:", func(t *testing.T) {
			chanSh := make(chan struct{})
			err := test.storage.Init(false, "", 0, chanSh, context.Background())
			if err != nil {
				panic(err)
			}
			test.storage.RepositoryAddCounterValue("PollCount", 1)
			test.storage.RepositoryAddGaugeValue("BuckHashSys", 0.1)

			request := httptest.NewRequest(http.MethodPost, test.request, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("metricType", test.metricType)
			rctx.URLParams.Add("metricName", test.metricName)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			logger, err := zap.NewDevelopment()
			if err != nil {
				panic(err)
			}

			defer logger.Sync()
			App := Application{Storage: test.storage, Logger: *logger.Sugar()}

			w := httptest.NewRecorder()

			h := http.HandlerFunc(App.GetMetricPath())
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

func TestUpdateAllValues(t *testing.T) {
	var gaugeMetricValue = 1.5

	type httpResult struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name    string
		request string
		metrics []data.Metrics
		storage *str.MemStorage
		result  httpResult
	}{
		{
			name:    "test: Add correct metrics",
			request: "/updates/",
			metrics: []data.Metrics{{ID: "GaugeMetric", MType: "gauge", Value: &gaugeMetricValue}},
			storage: &str.MemStorage{},
			result: httpResult{
				code:        200,
				contentType: "application/json",
			},
		},

		{
			name:    "test: Add metric without name",
			request: "/updates/",
			metrics: []data.Metrics{{ID: "", MType: "gauge", Value: &gaugeMetricValue}},
			storage: &str.MemStorage{},
			result: httpResult{
				code:        404,
				response:    "Error 404: Metric name was not found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},

		{
			name:    "test: Add metric with incorrect type",
			request: "/updates/",
			metrics: []data.Metrics{{ID: "GaugeMetric", MType: "test", Value: &gaugeMetricValue}},
			storage: &str.MemStorage{},
			result: httpResult{
				code:        400,
				response:    "Error 400: Invalid metric type: test\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, test := range tests {
		t.Run("Test:", func(t *testing.T) {
			chanSh := make(chan struct{})
			err := test.storage.Init(false, "", 0, chanSh, context.Background())
			if err != nil {
				panic(err)
			}

			var buf bytes.Buffer
			bodyRequestEncode := json.NewEncoder(&buf)
			err = bodyRequestEncode.Encode(test.metrics)
			if err != nil {
				panic(err)
			}

			request := httptest.NewRequest(http.MethodGet, test.request, &buf)
			logger, err := zap.NewDevelopment()
			if err != nil {
				panic(err)
			}

			defer logger.Sync()
			App := Application{Storage: test.storage, Logger: *logger.Sugar()}

			w := httptest.NewRecorder()

			h := http.HandlerFunc(App.MiddlewareZipper(App.UpdateAllValues()))

			h(w, request)

			res := w.Result()
			assert.Equal(t, test.result.code, res.StatusCode)

			defer res.Body.Close()
		})
	}
}

func BenchmarkGetMetricPath(b *testing.B) {
	storage := &str.MemStorage{}
	chanSh := make(chan struct{})
	err := storage.Init(false, "", 0, chanSh, context.Background())
	if err != nil {
		panic(err)
	}
	storage.RepositoryAddCounterValue("PollCount", 1)
	storage.RepositoryAddGaugeValue("BuckHashSys", 0.1)

	request := httptest.NewRequest("GET", "/value/", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("metricType", "counter")
	rctx.URLParams.Add("metricName", "PollCount")
	req := request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()
	App := Application{Storage: storage, Logger: *logger.Sugar()}

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(App.GetMetricPath())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkUpdateValue(b *testing.B) {
	storage := &str.MemStorage{}
	var counterMetrciValue int64 = 4
	metric := &data.Metrics{ID: "value", MType: "counter", Delta: &counterMetrciValue}
	var buf bytes.Buffer
	chanSh := make(chan struct{})
	err := storage.Init(false, "", 0, chanSh, context.Background())
	if err != nil {
		panic(err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()
	App := Application{Storage: storage, Logger: *logger.Sugar()}
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(App.UpdateValue())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bodyRequestEncode := json.NewEncoder(&buf)
		err = bodyRequestEncode.Encode(metric)
		if err != nil {
			panic(err)
		}

		request := httptest.NewRequest(http.MethodPost, "/update/", &buf)
		
		
		handler.ServeHTTP(w, request)
	}

}

func BenchmarkUpdateValuePath(b *testing.B) {
	storage := &str.MemStorage{}
	chanSh := make(chan struct{})
	err := storage.Init(false, "", 0, chanSh, context.Background())
	if err != nil {
		panic(err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()
	App := Application{Storage: storage, Logger: *logger.Sugar()}
	request := httptest.NewRequest("POST", "/value/", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("metricType", "counter")
	rctx.URLParams.Add("metricName", "PollCount")
	rctx.URLParams.Add("metricValue", "4")
	req := request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(App.UpdateValuePath())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkGetMetric(b *testing.B) {
	storage := &str.MemStorage{}
	chanSh := make(chan struct{})
	err := storage.Init(false, "", 0, chanSh, context.Background())
	if err != nil {
		panic(err)
	}
	storage.RepositoryAddCounterValue("PollCount", 1)
	storage.RepositoryAddGaugeValue("BuckHashSys", 0.1)

	metric := &data.Metrics{ID: "PollCount", MType: "counter"}

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()
	App := Application{Storage: storage, Logger: *logger.Sugar()}

	var buf bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bodyRequestEncode := json.NewEncoder(&buf)
		err = bodyRequestEncode.Encode(metric)
		if err != nil {
			panic(err)
		}

		request := httptest.NewRequest("POST", "/value/", &buf)

		w := httptest.NewRecorder()
		handler := http.HandlerFunc(App.GetMetric())

		handler.ServeHTTP(w, request)
	}
}
