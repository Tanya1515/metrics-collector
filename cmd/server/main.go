package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// func Split(s, sep string) []string - переписать, спросить у ментора, почему не работает PathValue

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

		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusOK)

		rw.Write([]byte("Succesfully edit!"))

	}
}

func main() {
	var Storage = &MemStorage{CounterStorage: make(map[string][]int64, 100), GaugeStorage: make(map[string]float64, 100)}
	mux := http.NewServeMux()
	// Handle registers the handler for the given pattern in Router/mux
	// otherwise HandleFunc registers the handler function for the given pattern in
	mux.HandleFunc(`/update/{metricType}/{metricName}/{metricValue}`, ProcessRequest(Storage))
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
