package main

import (
	"fmt"
	"net/http"
	"strconv"
)

var Storage = MemStorage{CounterStorage: make(map[string]int64, 1024), GaugeStorage: make(map[string]float64, 1024)}

func ValidateRequest(next http.Handler) http.Handler {

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(rw, "Error 405: only POST-requests are allowed", http.StatusMethodNotAllowed)
			return
		}

		metricType := r.PathValue("metricType")

		if (metricType != "counter") && (metricType != "gauge") {
			http.Error(rw, fmt.Sprintf("Error 400: Invalid metric type: %s", metricType), http.StatusBadRequest)
			return
		}

		metricName := r.PathValue("metricName")
		if metricName == "" {
			http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
			return
		}

		metricValue := r.PathValue("metricValue")

		if metricType == "counter" {
			_, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 400: Invalid metric value: %s", metricValue), http.StatusBadRequest)
				return
			}
		}
		if metricType == "gauge" {
			_, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 400: Invalid metric value: %s", metricValue), http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(rw, r)
	})
}

func ProcessMetrix(rw http.ResponseWriter, r *http.Request) {
	//save metrics
	metricType := r.PathValue("metricType")

	metricName := r.PathValue("metricName")

	metricValue := r.PathValue("metricValue")

	if metricType == "counter" {
		metricValueInt64, _ := strconv.ParseInt(metricValue, 10, 64)
		Storage.StorageAddCounterValue(metricName, metricValueInt64)
	}

	if metricType == "gauge" {
		metricValueFloat64, _ := strconv.ParseFloat(metricValue, 64)
		Storage.StorageAddGaugeValue(metricName, metricValueFloat64)
	}

	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)

	rw.Write([]byte("Succesfully edit"))
}

func main() {

	mux := http.NewServeMux()
	// Handle registers the handler for the given pattern in Router/mux
	// otherwise HandleFunc registers the handler function for the given pattern in
	mux.Handle(`/update/{metricType}/{metricName}/{metricValue}`, ValidateRequest(http.HandlerFunc(ProcessMetrix)))
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
