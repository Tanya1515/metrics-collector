package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"compress/gzip"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type ResultMetrics struct {
	GaugeMetrics   string
	CounterMetrics string
}

type Application struct {
	Storage *MemStorage
	logger  zap.SugaredLogger
}

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (App *Application) UpdateValue() http.HandlerFunc {
	updateValuefunc := func(rw http.ResponseWriter, r *http.Request) {
		var metricData Metrics
		var buf bytes.Buffer

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				App.logger.Errorln("Error during unpacking the request")
				return
			}
			defer gz.Close()

			_, err = buf.ReadFrom(gz)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

		} else {
			_, err := buf.ReadFrom(r.Body)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				App.logger.Errorln("Bad request catched")
				return
			}
		}
		if err := json.Unmarshal(buf.Bytes(), &metricData); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			App.logger.Errorln("Error during deserialization")
			return
		}

		if (metricData.MType != "counter") && (metricData.MType != "gauge") {
			http.Error(rw, fmt.Sprintf("Error 400: Invalid metric type: %s", metricData.MType), http.StatusBadRequest)
			App.logger.Errorln("Error 400: Invalid metric type:", metricData.MType)
			return
		}

		if metricData.ID == "" {
			http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
			App.logger.Errorln("Metric name was not found")
			return
		}

		if metricData.MType == "counter" {
			App.Storage.RepositoryAddCounterValue(metricData.ID, *metricData.Delta)
		}
		if metricData.MType == "gauge" {
			App.Storage.RepositoryAddGaugeValue(metricData.ID, *metricData.Value)
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)

		metricDataBytes, err := json.Marshal(metricData)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			App.logger.Errorln("Error during serialization")
		}
		rw.Write(metricDataBytes)

	}
	return http.HandlerFunc(updateValuefunc)
}

func (App *Application) HTMLMetrics() http.HandlerFunc {
	htmlMetricsfunc := func(rw http.ResponseWriter, r *http.Request) {

		builder := strings.Builder{}
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
		t, err := template.ParseFiles("./html/metrics.html")
		if err != nil {
			http.Error(rw, "Error 500: error while processing html page", http.StatusInternalServerError)
			App.logger.Errorln("Error while processing html page:", err)
		}
		t.Execute(rw, res)
	}

	return http.HandlerFunc(htmlMetricsfunc)
}

func (App *Application) GetMetric() http.HandlerFunc {
	getMetricfunc := func(rw http.ResponseWriter, r *http.Request) {
		metricData := Metrics{}
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			App.logger.Errorln("Bad request catched")
			return
		}
		if err = json.Unmarshal(buf.Bytes(), &metricData); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			App.logger.Errorln("Error during deserialization")
			return
		}
		if metricData.ID == "" {
			http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
			App.logger.Errorln("Metric name was not found")
			return
		}
		if metricData.MType == "counter" {
			metricValue, err := App.Storage.GetCounterValueByName(metricData.ID)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 404: %s", err), http.StatusNotFound)
				App.logger.Errorln("Error in CounterStorage:", err)
				return
			}
			metricData.Delta = &metricValue
		} else if metricData.MType == "gauge" {
			metricValue, err := App.Storage.GetGaugeValueByName(metricData.ID)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 404: %s", err), http.StatusNotFound)
				App.logger.Errorln("Error in GaugeStorage:", err)
				return
			}
			metricData.Value = &metricValue
		} else {
			http.Error(rw, fmt.Sprintf("Error 400: Invalid metric type: %s", metricData.MType), http.StatusBadRequest)
			App.logger.Errorln("Invalid metric type:", metricData.MType)
			return
		}

		metricDataBytes, err := json.Marshal(metricData)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			App.logger.Errorln("Error during serialization")
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write(metricDataBytes)

	}
	return http.HandlerFunc(getMetricfunc)
}

func (App *Application) WithLoggerZipper(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uri := r.RequestURI
		method := r.Method

		responseData := &ResponseData{
			status: 0,
			size:   0,
		}

		zlw := LoggingZipperResponseWriter{
			w,
			w,
			responseData,
		}
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && (strings.Contains(r.Header.Get("Content-Type"), "application/json") || strings.Contains(r.Header.Get("Content-Type"), "text/html")) {
			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				App.logger.Errorln("Error during comressing")
			}

			defer gz.Close()
			w.Header().Set("Content-Encoding", "gzip")
			zlw = LoggingZipperResponseWriter{
				w,
				gz,
				responseData,
			}
		}

		start := time.Now()

		h.ServeHTTP(&zlw, r)

		duration := time.Since(start)

		App.logger.Infoln(
			"URI", uri,
			"Method", method,
			"Duration", duration,
			"ResponseStatus", responseData.status,
			"ResponseSize", responseData.size,
		)

	}
}

func main() {
	var mutex sync.Mutex
	var Storage = &MemStorage{CounterStorage: make(map[string]int64, 100), GaugeStorage: make(map[string]float64, 100), mutex: &mutex}
	serverAddressFlag := flag.String("a", "localhost:8080", "server address")

	flag.Parse()

	serverAddress := "localhost:8080"

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	App := Application{Storage: Storage, logger: *logger.Sugar()}

	App.logger.Infow(
		"Starting server",
		"addr", serverAddress,
	)

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", App.HTMLMetrics())
		r.Get("/value", App.GetMetric())
		r.Post("/update", App.UpdateValue())
	})

	serverAddress, envExists := os.LookupEnv("ADDRESS")
	if !(envExists) {
		serverAddress = *serverAddressFlag
	}
	err = http.ListenAndServe(serverAddress, App.WithLoggerZipper(r))
	if err != nil {
		App.logger.Fatalw(err.Error(), "event", "start server")
	}
}
