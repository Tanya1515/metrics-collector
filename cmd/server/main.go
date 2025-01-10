package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
)

type ResultMetrics struct {
	GaugeMetrics   string
	CounterMetrics string
}

type Application struct {
	Storage *MemStorage
}

func (App *Application) ProcessRequest() http.HandlerFunc {

	return func(rw http.ResponseWriter, r *http.Request) {

		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		if (metricType != "counter") && (metricType != "gauge") {
			http.Error(rw, fmt.Sprintf("Error 400: Invalid metric type: %s", metricType), http.StatusBadRequest)
			return
		}

		if metricName == "" {
			http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
			return
		}
		App.Storage.mutex.Lock()
		if metricType == "counter" {
			metricValueInt64, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 400: Invalid metric value: %s", metricValue), http.StatusBadRequest)
				App.Storage.mutex.Unlock()
				return
			}
			App.Storage.RepositoryAddCounterValue(metricName, metricValueInt64)
		}
		if metricType == "gauge" {
			metricValueFloat64, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 400: Invalid metric value: %s", metricValue), http.StatusBadRequest)
				App.Storage.mutex.Unlock()
				return
			}
			App.Storage.RepositoryAddGaugeValue(metricName, metricValueFloat64)
		}
		App.Storage.mutex.Unlock()

		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.WriteHeader(http.StatusOK)

		rw.Write([]byte("Succesfully edit!"))

	}
}

func (App *Application) HTMLMetrics() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		builder := strings.Builder{}
		App.Storage.mutex.Lock()
		for key, value := range App.Storage.GaugeStorage {
			builder.WriteString(key)
			builder.WriteString(": ")
			builder.WriteString(strconv.FormatFloat(value, 'f', -1, 64))
			builder.WriteString(" \n")
		}
		gaugeResult := builder.String()

		builder = strings.Builder{}
		for key, value := range App.Storage.CounterStorage {
			builder.WriteString(key)
			builder.WriteString(": ")
			builder.WriteString(strconv.FormatInt(value, 10))
			builder.WriteString(" \n")
		}
		counterResult := builder.String()

		res := ResultMetrics{GaugeMetrics: gaugeResult, CounterMetrics: counterResult}
		App.Storage.mutex.Unlock()
		t, err := template.ParseFiles("./html/metrics.html")
		if err != nil {
			http.Error(rw, "Error 500: error while processing html page", http.StatusInternalServerError)
		}
		t.Execute(rw, res)
	}
}

func (App *Application) GetMetric() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		if metricName == "" {
			http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
			return
		}
		metricRes := ""

		App.Storage.mutex.Lock()
		if metricType == "counter" {
			metricValue, err := App.Storage.GetCounterValueByName(metricName)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 404: %s", err), http.StatusNotFound)
				App.Storage.mutex.Unlock()
				return
			}
			builder := strings.Builder{}
			builder.WriteString(strconv.FormatInt(metricValue, 10))
			metricRes = builder.String()
		} else if metricType == "gauge" {
			metricValue, err := App.Storage.GetGaugeValueByName(metricName)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 404: %s", err), http.StatusNotFound)
				App.Storage.mutex.Unlock()
				return
			}
			metricRes = strconv.FormatFloat(metricValue, 'f', -1, 64)
		} else {
			http.Error(rw, fmt.Sprintf("Error 400: Invalid metric type: %s", metricType), http.StatusBadRequest)
			App.Storage.mutex.Unlock()
			return
		}
		App.Storage.mutex.Unlock()

		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(metricRes))

	}
}

func main() {
	var mutex sync.Mutex
	var Storage = &MemStorage{CounterStorage: make(map[string]int64, 100), GaugeStorage: make(map[string]float64, 100), mutex: &mutex}
	App := Application{Storage: Storage}
	serverAddressFlag := flag.String("a", "localhost:8080", "server address")

	flag.Parse()

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", App.HTMLMetrics())
		r.Get("/value/{metricType}/{metricName}", App.GetMetric())
		r.Post("/update/{metricType}/{metricName}/{metricValue}", App.ProcessRequest())
	})

	serverAddress, envExists := os.LookupEnv("ADDRESS")
	if !(envExists) {
		serverAddress = *serverAddressFlag
	}
	fmt.Printf("Start server on address %s\n", serverAddress)
	err := http.ListenAndServe(serverAddress, r)
	if err != nil {
		panic(err)
	}
}
