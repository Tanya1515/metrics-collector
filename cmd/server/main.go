package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type ResultMetrics struct {
	GaugeMetrics   string
	CounterMetrics string
}

type Application struct{
	Storage *MemStorage
}

func (App *Application) ProcessRequest() http.HandlerFunc {

	return func(rw http.ResponseWriter, r *http.Request) {

		metrics := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/")

		if (metrics[0] != "counter") && (metrics[0] != "gauge") {
			http.Error(rw, fmt.Sprintf("Error 400: Invalid metric type: %s", metrics[0]), http.StatusBadRequest)
			return
		}

		if metrics[1] == "" {
			http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
			return
		}
		App.Storage.mutex.Lock()
		if metrics[0] == "counter" {
			metricValueInt64, err := strconv.ParseInt(metrics[2], 10, 64)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 400: Invalid metric value: %s", metrics[2]), http.StatusBadRequest)
				return
			}
			App.Storage.RepositoryAddCounterValue(metrics[1], metricValueInt64)
		}
		if metrics[0] == "gauge" {
			metricValueFloat64, err := strconv.ParseFloat(metrics[2], 64)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 400: Invalid metric value: %s", metrics[2]), http.StatusBadRequest)
				return
			}
			App.Storage.RepositoryAddGaugeValue(metrics[1], metricValueFloat64)
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
		metric := strings.Split(strings.TrimPrefix(r.URL.Path, "/value/"), "/")

		if metric[1] == "" {
			http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
			return
		}
		metricRes := ""

		App.Storage.mutex.Lock()
		if metric[0] == "counter" {
			metricValue, err := App.Storage.GetCounterValueByName(metric[1])
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 404: %s", err), http.StatusNotFound)
				return
			}
			builder := strings.Builder{}
			builder.WriteString(strconv.FormatInt(metricValue, 10))
			metricRes = builder.String()
		} else if metric[0] == "gauge" {
			metricValue, err := App.Storage.GetGaugeValueByName(metric[1])
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 404: %s", err), http.StatusNotFound)
				return
			}
			metricRes = strconv.FormatFloat(metricValue, 'f', -1, 64)
		} else {
			http.Error(rw, fmt.Sprintf("Error 400: Invalid metric type: %s", metric[0]), http.StatusBadRequest)
			return
		}
		App.Storage.mutex.Unlock()

		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(metricRes))
		
	}
}

func main() {
	var Storage = &MemStorage{CounterStorage: make(map[string]int64, 100), GaugeStorage: make(map[string]float64, 100)}
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
	err := http.ListenAndServe(serverAddress, r)
	if err != nil {
		panic(err)
	}
}
