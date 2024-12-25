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
	"time"

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

func (App *Application) UpdateValue() http.Handler {
	updateValuefunc := func(rw http.ResponseWriter, r *http.Request) {

		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		if (metrics[0] != "counter") && (metrics[0] != "gauge") {
			http.Error(rw, fmt.Sprintf("Error 400: Invalid metric type: %s", metrics[0]), http.StatusBadRequest)
			sugar.Errorln("Expected Post method, but recieved:", r.Method)
			return
		}

		if metricName == "" {
			http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
			sugar.Errorln("Metric name was not found")
			return
		}
		if metricType == "counter" {
			metricValueInt64, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 400: Invalid metric value: %s", metrics[2]), http.StatusBadRequest)
				sugar.Errorln("Invalid metric value:", err)
				return
			}
			App.Storage.RepositoryAddCounterValue(metricName, metricValueInt64)
		}
		if metricType == "gauge" {
			metricValueFloat64, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 400: Invalid metric value: %s", metrics[2]), http.StatusBadRequest)
				sugar.Errorln("Invalid metric value:", err)
				return
			}
			App.Storage.RepositoryAddGaugeValue(metricName, metricValueFloat64)
		}

		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.WriteHeader(http.StatusOK)

		rw.Write([]byte("Succesfully edit!"))

	}
	return http.HandlerFunc(updateValuefunc)
}

func (App *Application) HTMLMetrics() http.Handler {
	htmlMetricsfunc := func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(rw, "Error 405: only GET-requests are allowed", http.StatusMethodNotAllowed)
			sugar.Errorln("Expected Get method, but recieved:", r.Method)
			return
		}
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
			sugar.Errorln("Error while processing html page:", err)
		}
		t.Execute(rw, res)
	}

	return http.HandlerFunc(htmlMetricsfunc)
}

func (App *Application) GetMetric() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		if metricName == "" {
			http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
			App.logger.Errorln("Metric name was not found")
			return
		}
		metricRes := ""

		if metricType == "counter" {
			metricValue, err := App.Storage.GetCounterValueByName(metricName)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 404: %s", err), http.StatusNotFound)
				sugar.Errorln("Error in CounterStorage:", err)
				return
			}
			builder := strings.Builder{}
			builder.WriteString(strconv.FormatInt(metricValue, 10))
			metricRes = builder.String()
		} else if metricType == "gauge" {
			metricValue, err := App.Storage.GetGaugeValueByName(metricName)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 404: %s", err), http.StatusNotFound)
				sugar.Errorln("Error in GaugeStorage:", err)
				return
			}
			metricRes = strconv.FormatFloat(metricValue, 'f', -1, 64)
		} else {
			http.Error(rw, fmt.Sprintf("Error 400: Invalid metric type: %s", metric[0]), http.StatusBadRequest)
			sugar.Errorln("Invalid metric type:", metric[0])
			return
		}

		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(metricRes))

	}
}

func (App *Application) WithLogger(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		responseData := &ResponseData{
			status: 0,
			size:   0,
		}

		lw := LoggingResponseWriter{
			w,
			responseData,
		}

		h.ServeHTTP(&lw, r)

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
		r.Get("/", App.WithLogger(App.HTMLMetrics()))
		r.Get("/value/{metricType}/{metricName}", App.WithLogger(App.GetMetric()))
		r.Post("/update/{metricType}/{metricName}/{metricValue}", App.WithLogger(App.UpdateValue()))
	})

	serverAddress, envExists := os.LookupEnv("ADDRESS")
	if !(envExists) {
		serverAddress = *serverAddressFlag
	}
	fmt.Printf("Start server on address %s\n", serverAddress)
	err = http.ListenAndServe(serverAddress, r)
	if err != nil {
		App.logger.Fatalw(err.Error(), "event", "start server")
	}
}
