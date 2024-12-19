package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	//"html/template"

	"github.com/go-chi/chi/v5"
)

func ProcessRequest(Storage *MemStorage) http.HandlerFunc {

	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(rw, "Error 405: only POST-requests are allowed", http.StatusMethodNotAllowed)
			return
		}

		metrics := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/")

		if (metrics[0] != "counter") && (metrics[0] != "gauge") {
			http.Error(rw, fmt.Sprintf("Error 400: Invalid metric type: %s", metrics[0]), http.StatusBadRequest)
			return
		}

		if metrics[1] == "" {
			http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
			return
		}

		if metrics[0] == "counter" {
			metricValueInt64, err := strconv.ParseInt(metrics[2], 10, 64)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 400: Invalid metric value: %s", metrics[2]), http.StatusBadRequest)
				return
			}
			Storage.RepositoryAddCounterValue(metrics[1], metricValueInt64)
		}
		if metrics[0] == "gauge" {
			metricValueFloat64, err := strconv.ParseFloat(metrics[2], 64)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 400: Invalid metric value: %s", metrics[2]), http.StatusBadRequest)
				return
			}
			Storage.RepositoryAddGaugeValue(metrics[1], metricValueFloat64)
		}

		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.WriteHeader(http.StatusOK)

		rw.Write([]byte("Succesfully edit!"))

	}
}

func HTMLMetrics(Storage *MemStorage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(rw, "Error 405: only GET-requests are allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func GetMetric(Storage *MemStorage) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(rw, "Error 405: only GET-requests are allowed", http.StatusMethodNotAllowed)
			return
		}
		metric := strings.Split(strings.TrimPrefix(r.URL.Path, "/value/"), "/")

		if metric[1] == "" {
			http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
			return
		}

		if metric[0] == "counter" {
			metricValue, err := Storage.GetCounterValueByName(metric[1])
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 404: %s", err), http.StatusNotFound)
				return
			}
			builder := strings.Builder{}
			for _, value := range metricValue {
				builder.WriteString(strconv.FormatInt(value, 10))
			}

			rw.Write([]byte(builder.String()))
		} else if metric[0] == "gauge" {
			metricValue, err := Storage.GetGaugeValueByName(metric[1])
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 404: %s", err), http.StatusNotFound)
				return
			}
			rw.Write([]byte(fmt.Sprintf("%f", metricValue)))
		} else {
			http.Error(rw, fmt.Sprintf("Error 400: Invalid metric type: %s", metric[0]), http.StatusBadRequest)
			return
		}

		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.WriteHeader(http.StatusOK)
	}
}

func main() {
	var Storage = &MemStorage{CounterStorage: make(map[string][]int64, 100), GaugeStorage: make(map[string]float64, 100)}

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", HTMLMetrics(Storage))
		r.Get("/value/{metricType}/{metricName}", GetMetric(Storage))
		r.Post("/update/{metricType}/{metricName}/{metricValue}", ProcessRequest(Storage))
	})
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
