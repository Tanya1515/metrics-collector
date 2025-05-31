package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
	str "github.com/Tanya1515/metrics-collector.git/cmd/storage/structure"
)

func ExampleApplication_GetMetricPath() {
	storage := &str.MemStorage{}
	chanSh := make(chan struct{})
	storage.Init(context.Background(), chanSh)
	storage.RepositoryAddCounterValue("PollCount", 1)
	storage.RepositoryAddGaugeValue("BuckHashSys", 0.1)

	request := httptest.NewRequest(http.MethodPost, "/value/", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("metricType", "gauge")
	rctx.URLParams.Add("metricName", "BuckHashSys")
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()
	App := Application{Storage: storage, Logger: *logger.Sugar()}

	w := httptest.NewRecorder()

	h := http.HandlerFunc(App.GetMetricPath())
	h(w, request)

	res := w.Result()

	resBody, err := io.ReadAll(res.Body)

	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	fmt.Println(res.StatusCode)
	fmt.Println(string(resBody))

	// Output:
	// 200
	// 0.1
}

func ExampleApplication_UpdateValuePath() {
	storage := &str.MemStorage{}
	chanSh := make(chan struct{})
	storage.Init(context.Background(), chanSh)
	storage.RepositoryAddCounterValue("PollCount", 1)
	storage.RepositoryAddGaugeValue("BuckHashSys", 0.1)

	request := httptest.NewRequest(http.MethodPost, "/update", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("metricType", "counter")
	rctx.URLParams.Add("metricName", "TestCounter")
	rctx.URLParams.Add("metricValue", "10")
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()
	App := Application{Storage: storage, Logger: *logger.Sugar()}

	w := httptest.NewRecorder()

	h := http.HandlerFunc(App.UpdateValuePath())
	h(w, request)

	res := w.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)

	value, _ := storage.GetCounterValueByName("TestCounter")
	fmt.Println(value)

	// Output:
	// 200
	// 10

}

func ExampleApplication_GetMetric() {
	storage := &str.MemStorage{}
	chanSh := make(chan struct{})
	storage.Init(context.Background(), chanSh)
	storage.RepositoryAddCounterValue("PollCount", 1)
	storage.RepositoryAddGaugeValue("BuckHashSys", 0.1)

	metric := &data.Metrics{ID: "PollCount", MType: "counter"}
	var buf bytes.Buffer
	bodyRequestEncode := json.NewEncoder(&buf)
	err := bodyRequestEncode.Encode(metric)
	if err != nil {
		panic(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/value", &buf)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()
	App := Application{Storage: storage, Logger: *logger.Sugar()}

	w := httptest.NewRecorder()

	h := http.HandlerFunc(App.GetMetric())
	h(w, request)

	res := w.Result()

	fmt.Println(res.StatusCode)

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	fmt.Println(string(resBody))

	// Output:
	// 200
	// {"id":"PollCount","type":"counter","delta":1}

}

func ExampleApplication_UpdateValue() {
	storage := &str.MemStorage{}
	var counterMetrciValue int64 = 4
	chanSh := make(chan struct{})
	storage.Init(context.Background(), chanSh)
	storage.RepositoryAddCounterValue("PollCount", 1)
	storage.RepositoryAddGaugeValue("BuckHashSys", 0.1)

	var buf bytes.Buffer
	bodyRequestEncode := json.NewEncoder(&buf)
	metric := &data.Metrics{ID: "PollCount", MType: "counter", Delta: &counterMetrciValue}
	err := bodyRequestEncode.Encode(metric)
	if err != nil {
		panic(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/update/", &buf)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()
	App := Application{Storage: storage, Logger: *logger.Sugar()}

	w := httptest.NewRecorder()

	h := http.HandlerFunc(App.UpdateValue())
	h(w, request)

	res := w.Result()

	fmt.Println(res.StatusCode)

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	fmt.Println(string(resBody))

	delta, _ := storage.GetCounterValueByName("PollCount")

	fmt.Println(delta)

	// Output:
	// 200
	// {"id":"PollCount","type":"counter","delta":4}
	// 5
}

func ExampleApplication_UpdateAllValues() {
	storage := &str.MemStorage{}
	chanSh := make(chan struct{})
	storage.Init(context.Background(), chanSh)
	storage.RepositoryAddCounterValue("PollCount", 1)
	storage.RepositoryAddGaugeValue("BuckHashSys", 0.1)

	metrics := make([]data.Metrics, 2)
	var testCounterAllDelta int64 = 101
	testGaugeAllValue := 101.101
	metrics[0] = data.Metrics{ID: "TestCounterAll", MType: "counter", Delta: &testCounterAllDelta}
	metrics[1] = data.Metrics{ID: "TestGaugeAll", MType: "gauge", Value: &testGaugeAllValue}

	var buf bytes.Buffer
	bodyRequestEncode := json.NewEncoder(&buf)
	err := bodyRequestEncode.Encode(metrics)
	if err != nil {
		panic(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/updates/", &buf)
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()
	App := Application{Storage: storage, Logger: *logger.Sugar()}

	w := httptest.NewRecorder()

	h := http.HandlerFunc(App.UpdateAllValues())

	h(w, request)

	res := w.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)

	delta, _ := storage.GetCounterValueByName("TestCounterAll")

	fmt.Println(delta)

	value, _ := storage.GetGaugeValueByName("TestGaugeAll")

	fmt.Println(value)

	// Output:
	// 200
	// 101
	// 101.101
}
