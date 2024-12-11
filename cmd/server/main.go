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

		metric_type := r.PathValue("metric_type")

		if (metric_type != "counter") && (metric_type != "gauge") {
			http.Error(rw, fmt.Sprintf("Error 400: Invalid metric type: %s", metric_type), http.StatusBadRequest)
			return
		}

		metric_name := r.PathValue("metric_name")
		if metric_name == "" {
			http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
			return
		}

		metric_value := r.PathValue("metric_value")

		if metric_type == "counter" {
			_, err := strconv.ParseInt(metric_value, 10, 64)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 400: Invalid metric value: %s", metric_value), http.StatusBadRequest)
				return
			}
		}
		if metric_type == "gauge" {
			_, err := strconv.ParseFloat(metric_value, 64)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 400: Invalid metric value: %s", metric_value), http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(rw, r)
	})
}

func ProcessMetrix(rw http.ResponseWriter, r *http.Request) {
	//save metrics
	metric_type := r.PathValue("metric_type")

	metric_name := r.PathValue("metric_name")

	metric_value := r.PathValue("metric_value")

	if metric_type == "counter" {
		metric_value_int64, _ := strconv.ParseInt(metric_value, 10, 64)
		Storage.StorageAddCounterValue(metric_name, metric_value_int64)
	}

	if metric_type == "gauge" {
		metric_value_float64, _ := strconv.ParseFloat(metric_value, 64)
		Storage.StorageAddGaugeValue(metric_name, metric_value_float64)
	}

	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusOK)

	rw.Write([]byte("Succesfully edit"))

	return
}

func main() {

	mux := http.NewServeMux()
	// Handle registers the handler for the given pattern in Router/mux
	// otherwise HandleFunc registers the handler function for the given pattern in
	mux.Handle(`/update/{metric_type}/{metric_name}/{metric_value}`, ValidateRequest(http.HandlerFunc(ProcessMetrix)))
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
