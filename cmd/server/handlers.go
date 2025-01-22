package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"compress/gzip"
)

func (App *Application) UpdateValue() http.HandlerFunc {
	updateValuefunc := func(rw http.ResponseWriter, r *http.Request) {
		var metricData Metrics
		var buf bytes.Buffer
		var err error

		metricData.MType = chi.URLParam(r, "metricType")
		metricData.ID = chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		if metricData.ID == "" {
			if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				gz, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
					App.Logger.Errorln("Error during unpacking the request")
					return
				}
				defer gz.Close()

				_, err = buf.ReadFrom(gz)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
					return
				}

			} else {
				_, err = buf.ReadFrom(r.Body)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusBadRequest)
					App.Logger.Errorln("Bad request catched")
					return
				}
			}
			if err := json.Unmarshal(buf.Bytes(), &metricData); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				App.Logger.Errorln("Error during deserialization")
				return
			}
		}

		if (metricData.MType != "counter") && (metricData.MType != "gauge") {
			http.Error(rw, fmt.Sprintf("Error 400: Invalid metric type: %s", metricData.MType), http.StatusBadRequest)
			App.Logger.Errorln("Error 400: Invalid metric type:", metricData.MType)
			return
		}

		if metricData.ID == "" {
			http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
			App.Logger.Errorln("Metric name was not found")
			return
		}
		if metricData.MType == "counter" {
			if metricValue != "" {
				metricValueInt64, err := strconv.ParseInt(metricValue, 10, 64)
				if err != nil {
					http.Error(rw, fmt.Sprintf("Error 400: Invalid metric value: %s", metricValue), http.StatusBadRequest)
					return
				}
				metricData.Delta = &metricValueInt64
			}

			App.Storage.RepositoryAddCounterValue(metricData.ID, *metricData.Delta)
		}
		if metricData.MType == "gauge" {
			if metricValue != "" {
				metricValueFloat64, err := strconv.ParseFloat(metricValue, 64)
				if err != nil {
					http.Error(rw, fmt.Sprintf("Error 400: Invalid metric value: %s", metricValue), http.StatusBadRequest)
					return
				}
				metricData.Value = &metricValueFloat64
			}
			App.Storage.RepositoryAddGaugeValue(metricData.ID, *metricData.Value)
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)

		metricDataBytes, err := json.Marshal(metricData)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			App.Logger.Errorln("Error during serialization")
		}

		if App.Backup {
			file, err := os.OpenFile(App.FileStore, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				App.Logger.Errorln("Error while openning file for backup")
			}

			defer file.Close()

			_, err = file.Write(metricDataBytes)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				App.Logger.Errorln("Error while writting data for backup")
			}
			_, err = file.WriteString("\n")
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				App.Logger.Errorln("Error while writting line transition: %s", err)
			}
		}

		rw.Write(metricDataBytes)

	}
	return http.HandlerFunc(updateValuefunc)
}

func (App *Application) HTMLMetrics() http.HandlerFunc {
	htmlMetricsfunc := func(rw http.ResponseWriter, r *http.Request) {

		builder := strings.Builder{}
		allGaugeMetrics := App.Storage.GetAllGaugeMetrics()
		for key, value := range allGaugeMetrics {
			builder.WriteString(key)
			builder.WriteString(": ")
			builder.WriteString(strconv.FormatFloat(value, 'f', -1, 64))
			builder.WriteString(" \n")
		}
		gaugeResult := builder.String()

		builder = strings.Builder{}
		allCounterMetrics := App.Storage.GetAllCounterMetrics()
		for key, value := range allCounterMetrics {
			builder.WriteString(key)
			builder.WriteString(": ")
			builder.WriteString(strconv.FormatInt(value, 10))
			builder.WriteString(" \n")
		}
		counterResult := builder.String()

		res := ResultMetrics{GaugeMetrics: gaugeResult, CounterMetrics: counterResult}
		tmpl := template.Must(template.New("template").Parse(`
Counter metrics: 

{{.CounterMetrics}}

Gauge metrics:

{{.GaugeMetrics}}

		`))

		tmpl.Execute(rw, res)
	}

	return http.HandlerFunc(htmlMetricsfunc)
}

func (App *Application) GetMetricPath() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		if metricName == "" {
			http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
			App.Logger.Errorln("Metric name was not found")
			return
		}
		metricRes := ""

		if metricType == "counter" {
			metricValue, err := App.Storage.GetCounterValueByName(metricName)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 404: %s", err), http.StatusNotFound)
				App.Logger.Errorln("Error in CounterStorage:", err)
				return
			}
			builder := strings.Builder{}
			builder.WriteString(strconv.FormatInt(metricValue, 10))
			metricRes = builder.String()
		} else if metricType == "gauge" {
			metricValue, err := App.Storage.GetGaugeValueByName(metricName)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 404: %s", err), http.StatusNotFound)
				App.Logger.Errorln("Error in GaugeStorage:", err)
				return
			}
			metricRes = strconv.FormatFloat(metricValue, 'f', -1, 64)
		} else {
			http.Error(rw, fmt.Sprintf("Error 400: Invalid metric type: %s", metricType), http.StatusBadRequest)
			App.Logger.Errorln("Invalid metric type:", metricType)
			return
		}

		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(metricRes))

	}
}

func (App *Application) GetMetric() http.HandlerFunc {
	getMetricfunc := func(rw http.ResponseWriter, r *http.Request) {
		metricData := Metrics{}

		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			App.Logger.Errorln("Bad request catched")
			return
		}
		if err = json.Unmarshal(buf.Bytes(), &metricData); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			App.Logger.Errorln("Error during deserialization")
			return
		}

		if metricData.ID == "" {
			http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
			App.Logger.Errorln("Metric name was not found")
			return
		}
		if metricData.MType == "counter" {
			metricValue, err := App.Storage.GetCounterValueByName(metricData.ID)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 404: %s", err), http.StatusNotFound)
				App.Logger.Errorln("Error in CounterStorage:", err)
				return
			}
			metricData.Delta = &metricValue
		} else if metricData.MType == "gauge" {
			metricValue, err := App.Storage.GetGaugeValueByName(metricData.ID)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 404: %s", err), http.StatusNotFound)
				App.Logger.Errorln("Error in GaugeStorage:", err)
				return
			}
			metricData.Value = &metricValue
		} else {
			http.Error(rw, fmt.Sprintf("Error 400: Invalid metric type: %s", metricData.MType), http.StatusBadRequest)
			App.Logger.Errorln("Invalid metric type:", metricData.MType)
			return
		}

		metricDataBytes, err := json.Marshal(metricData)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			App.Logger.Errorln("Error during serialization")
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
				App.Logger.Errorln("Error during comressing")
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

		App.Logger.Infoln(
			"URI", uri,
			"Method", method,
			"Duration", duration,
			"ResponseStatus", responseData.status,
			"ResponseSize", responseData.size,
		)

	}
}
