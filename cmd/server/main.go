package main

import (
	"bufio"
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

var (
	serverAddressFlag *string
	storeIntervalFlag *int
	fileStorePathFlag *string
	restoreFlag       *bool
)

func init() {
	serverAddressFlag = flag.String("a", "localhost:8080", "server address")
	storeIntervalFlag = flag.Int("i", 300, "time duration for saving metrics")
	fileStorePathFlag = flag.String("f", "/tmp/metrics-db.json", "filename for storing metrics")
	restoreFlag = flag.Bool("r", true, "store all info")
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

		if App.Storage.backup {
			file, err := os.OpenFile(App.Storage.fileStore, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				App.logger.Errorln("Error while openning file for backup")
			}

			defer file.Close()

			_, err = file.Write(metricDataBytes)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				App.logger.Errorln("Error while writting data for backup")
			}
			_, err = file.WriteString("\n")
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				App.logger.Errorln("Error while writting line transition: %s", err)
			}
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

func (App *Application) Store() error {

	file, err := os.OpenFile(App.Storage.fileStore, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		App.logger.Errorln("Error while openning file: %s", err)
		return err
	}

	scanner := bufio.NewScanner(file)

	defer file.Close()

	for {
		if !scanner.Scan() {
			return scanner.Err()
		}

		data := scanner.Bytes()

		metric := Metrics{}
		err = json.Unmarshal(data, &metric)
		if err != nil {
			App.logger.Errorln("Error while metric deserialization: ", err)
			return err
		}

		if metric.MType == "gauge" {
			App.Storage.RepositoryAddGaugeValue(metric.ID, *metric.Value)
		}

		if metric.MType == "counter" {
			App.Storage.RepositoryAddValue(metric.ID, *metric.Delta)
		}
	}
}

func (App *Application) SaveMetrics(timer time.Duration) {

	gaugeMetric := Metrics{ID: "", MType: "gauge"}
	counterMetric := Metrics{ID: "", MType: "counter"}
	for {
		App.logger.Infoln("Write data to backup file")
		file, err := os.OpenFile(App.Storage.fileStore, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			App.logger.Errorln("Error while openning file: %s", err)
		}
		App.Storage.mutex.Lock()
		for metricName, metricValue := range App.Storage.GaugeStorage {
			gaugeMetric.ID = metricName
			gaugeMetric.Value = &metricValue

			metricBytes, err := json.Marshal(gaugeMetric)
			if err != nil {
				App.logger.Errorln("Error while marshalling GaugeMetric: %s", err)
			}
			_, err = file.Write(metricBytes)
			if err != nil {
				App.logger.Errorln("Error while writing metric info to file: %s", err)
			}
			_, err = file.WriteString("\n")
			if err != nil {
				App.logger.Errorln("Error while writting line transition: %s", err)
			}
		}

		for metricName, metricValue := range App.Storage.CounterStorage {
			counterMetric.ID = metricName
			counterMetric.Delta = &metricValue

			metricBytes, err := json.Marshal(counterMetric)
			if err != nil {
				App.logger.Errorln("Error while marshalling CounterMetric: %s", err)
			}
			_, err = file.Write(metricBytes)
			if err != nil {
				App.logger.Errorln("Error while writing metric info to file: %s", err)
			}
			_, err = file.WriteString("\n")
			if err != nil {
				App.logger.Errorln("Error while writting line transition: %s", err)
			}
		}
		err = file.Close()
		if err != nil {
			App.logger.Errorln("Error while closing file: %s", err)
		}
		App.Storage.mutex.Unlock()

		time.Sleep(timer * time.Second)
	}

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
	var Storage = &MemStorage{CounterStorage: make(map[string]int64, 100), GaugeStorage: make(map[string]float64, 100), mutex: &mutex, backup: false, fileStore: ""}

	flag.Parse()

	serverAddress, envExists := os.LookupEnv("ADDRESS")
	if !(envExists) {
		serverAddress = *serverAddressFlag
	}

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
	storeInterval := 300
	restore := true
	storeIntervalEnv, envExists := os.LookupEnv("STORE_INTERVAL")
	if !(envExists) {
		storeInterval = *storeIntervalFlag
	} else {
		storeInterval, err = strconv.Atoi(storeIntervalEnv)
		if err != nil {
			App.logger.Errorln("Error when converting string to int: %v", err)
		}
	}

	App.Storage.fileStore, envExists = os.LookupEnv("FILE_STORAGE_PATH")
	if !(envExists) {
		App.Storage.fileStore = *fileStorePathFlag
	}

	restoreEnv, envExists := os.LookupEnv("RESTORE")
	if !(envExists) {
		restore = *restoreFlag
	} else {
		restore, err = strconv.ParseBool(restoreEnv)
		if err != nil {
			App.logger.Errorln("Error when converting string to bool: %s", err)
		}
	}
	if restore && (App.Storage.fileStore != "") {
		App.Store()
	}

	if (storeInterval != 0) && (App.Storage.fileStore != "") {
		go App.SaveMetrics(time.Duration(storeInterval))
	} else if (storeInterval == 0) && (App.Storage.fileStore != "") {
		App.Storage.backup = true
	}
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", App.HTMLMetrics())
		r.Get("/value", App.GetMetric())
		r.Post("/update", App.UpdateValue())
	})

	err = http.ListenAndServe(serverAddress, App.WithLoggerZipper(r))
	if err != nil {
		App.logger.Fatalw(err.Error(), "event", "start server")
	}
}
