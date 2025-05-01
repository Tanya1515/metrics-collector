package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
	retryerr "github.com/Tanya1515/metrics-collector.git/cmd/errors"
)

// UpdateValuePath - handler, that updates metric in PostgreSQL or in-memory storage.
// The function gets values from http-request as {metricType}/{metricName}/{metricValue}.
func (App *Application) UpdateValuePath() http.HandlerFunc {
	updateValuefunc := func(rw http.ResponseWriter, r *http.Request) {
		var metricData data.Metrics

		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		if (metricType != "counter") && (metricType != "gauge") {
			http.Error(rw, fmt.Sprintf("Error 400: Invalid metric type: %s", metricType), http.StatusBadRequest)
			App.Logger.Errorln("Invalid metric type:", metricType)
			return
		}

		if metricName == "" {
			http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
			App.Logger.Errorln("Metric name was not found")
			return
		}

		if metricType == "counter" {
			metricData.MType = "counter"
			metricValueInt64, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 400: Invalid metric value: %s", metricValue), http.StatusBadRequest)
				App.Logger.Errorln("Invalid metric value:", err)
				return
			}

			for i := 0; i <= 3; i++ {
				err = App.Storage.RepositoryAddCounterValue(metricName, metricValueInt64)
				if err == nil {
					break
				}
				if !(retryerr.CheckErrorType(err)) || (i == 3) {
					http.Error(rw, fmt.Sprintf("Error 500: Error while adding counter metric %s to Storage", metricData.ID), http.StatusInternalServerError)
					App.Logger.Errorln("Error while adding counter metric to Storage:", err)
					return
				}
				if i == 0 {
					time.Sleep(1 * time.Second)
				} else {
					time.Sleep(time.Duration(i+i+1) * time.Second)
				}
			}
		}
		if metricType == "gauge" {
			metricData.MType = "gauge"
			metricValueFloat64, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(rw, fmt.Sprintf("Error 400: Invalid metric value: %s", metricValue), http.StatusBadRequest)
				App.Logger.Errorln("Invalid metric value:", err)
				return
			}
			for i := 0; i <= 3; i++ {
				err = App.Storage.RepositoryAddGaugeValue(metricName, metricValueFloat64)
				if err == nil {
					break
				}
				if !(retryerr.CheckErrorType(err)) || (i == 3) {
					http.Error(rw, fmt.Sprintf("Error 500: Error while adding gauge metric %s to Storage", metricName), http.StatusInternalServerError)
					App.Logger.Errorln("Error while adding gauge metric to Storage:", err)
					return
				}
				if i == 0 {
					time.Sleep(1 * time.Second)
				} else {
					time.Sleep(time.Duration(i+i+1) * time.Second)
				}
			}
		}

		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.WriteHeader(http.StatusOK)

		rw.Write([]byte("Succesfully edit!"))

	}

	return http.HandlerFunc(updateValuefunc)
}

// UpdateValue - handler, that updates metric in PostgreSQL or in-memory storage.
// The function gets all data from request body.
func (App *Application) UpdateValue() http.HandlerFunc {
	updateValuefunc := func(rw http.ResponseWriter, r *http.Request) {
		var metricData data.Metrics
		var buf bytes.Buffer

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				App.Logger.Errorln("Error during unpacking the request: ", err)
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
				App.Logger.Errorln("Bad request catched")
				return
			}
		}

		if strings.Contains(r.Header.Get("X-Encrypted"), "rsa") {
			data.DecryptData(App.CryptoKey, buf.Bytes())
		}

		if err := json.Unmarshal(buf.Bytes(), &metricData); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			App.Logger.Errorln("Error during deserialization")
			return
		}

		if (metricData.MType != "counter") && (metricData.MType != "gauge") {
			http.Error(rw, fmt.Sprintf("Error 400: Invalid metric type: %s", metricData.MType), http.StatusBadRequest)
			App.Logger.Errorln("Error 400: Invalid metric type: ", metricData.MType)
			return
		}

		if metricData.ID == "" {
			http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
			App.Logger.Errorln("Metric name was not found")
			return
		}

		if metricData.MType == "counter" {
			for i := 0; i <= 3; i++ {
				err := App.Storage.RepositoryAddCounterValue(metricData.ID, *metricData.Delta)
				if err == nil {
					break
				}
				if !(retryerr.CheckErrorType(err)) || (i == 3) {
					http.Error(rw, fmt.Sprintf("Error 500: Error while adding counter metric %s to Storage", metricData.ID), http.StatusInternalServerError)
					App.Logger.Errorln("Error while adding counter metric to Storage:", err)
					return
				}
				if i == 0 {
					time.Sleep(1 * time.Second)
				} else {
					time.Sleep(time.Duration(i+i+1) * time.Second)
				}
			}
		}
		if metricData.MType == "gauge" {
			for i := 0; i <= 3; i++ {
				err := App.Storage.RepositoryAddGaugeValue(metricData.ID, *metricData.Value)
				if err == nil {
					break
				} else if !(retryerr.CheckErrorType(err)) || (i == 3) {
					http.Error(rw, fmt.Sprintf("Error 500: Error while adding gauge metric %s to Storage", metricData.ID), http.StatusInternalServerError)
					App.Logger.Errorln("Error while adding gauge metric to Storage:", err)
					return
				}
				if i == 0 {
					time.Sleep(1 * time.Second)
				} else {
					time.Sleep(time.Duration(i+i+1) * time.Second)
				}
			}
		}

		metricDataBytes, err := json.Marshal(metricData)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			App.Logger.Errorln("Error during serialization")
		}
		if App.SecretKey != "" {
			h := hmac.New(sha256.New, []byte(App.SecretKey))
			h.Write(metricDataBytes)
			signCheck := h.Sum(nil)
			rw.Header().Set("HashSHA256", hex.EncodeToString(signCheck))
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)

		rw.Write(metricDataBytes)

	}
	return http.HandlerFunc(updateValuefunc)
}

// HTMLMetrics - handler, that processes metrics from PostgreSQL or in-memory storage and display them in html-formet.
func (App *Application) HTMLMetrics() http.HandlerFunc {
	htmlMetricsfunc := func(rw http.ResponseWriter, r *http.Request) {

		builder := strings.Builder{}
		var allGaugeMetrics map[string]float64
		var err error
		for i := 0; i <= 3; i++ {
			allGaugeMetrics, err = App.Storage.GetAllGaugeMetrics()
			if err == nil {
				break
			}
			if !(retryerr.CheckErrorType(err)) || (i == 3) {
				http.Error(rw, "Error 500: Error while getting all gauge metrics", http.StatusInternalServerError)
				App.Logger.Errorln(err)
				return
			}
			if i == 0 {
				time.Sleep(1 * time.Second)
			} else {
				time.Sleep(time.Duration(i+i+1) * time.Second)
			}
		}
		for key, value := range allGaugeMetrics {
			builder.WriteString(key)
			builder.WriteString(": ")
			builder.WriteString(strconv.FormatFloat(value, 'f', -1, 64))
			builder.WriteString(" \n")
		}
		gaugeResult := builder.String()

		builder = strings.Builder{}
		var allCounterMetrics map[string]int64
		for i := 0; i <= 3; i++ {
			allCounterMetrics, err = App.Storage.GetAllCounterMetrics()
			if err == nil {
				break
			} else if !(retryerr.CheckErrorType(err)) {
				http.Error(rw, "Error 500: Error while getting all counter metrics", http.StatusInternalServerError)
				App.Logger.Errorln(err)
				return
			} else if retryerr.CheckErrorType(err) {
				if i == 0 {
					time.Sleep(1 * time.Second)
				} else {
					time.Sleep(time.Duration(i+i+1) * time.Second)
				}
			}
		}

		for key, value := range allCounterMetrics {
			builder.WriteString(key)
			builder.WriteString(": ")
			builder.WriteString(strconv.FormatInt(value, 10))
			builder.WriteString(" \n")
		}
		counterResult := builder.String()

		res := data.ResultMetrics{GaugeMetrics: gaugeResult, CounterMetrics: counterResult}
		tmpl := template.Must(template.New("template").Parse(`
Counter metrics: 

{{.CounterMetrics}}

Gauge metrics:

{{.GaugeMetrics}}

		`))
		rw.Header().Set("Content-Type", "text/html")
		rw.WriteHeader(http.StatusOK)
		tmpl.Execute(rw, res)
	}

	return http.HandlerFunc(htmlMetricsfunc)
}

// GetMetricPath - handler, that retrieve metrics value from PostgreSQL or in-memory storage and return the value.
// The function gets all metric type and name from URL-path.
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
		var err error
		if metricType == "counter" {
			var metricValue int64
			for i := 0; i <= 3; i++ {
				metricValue, err = App.Storage.GetCounterValueByName(metricName)
				if err == nil {
					break
				} else if !(retryerr.CheckErrorType(err)) {
					http.Error(rw, fmt.Sprintf("Error 404: %s", err), http.StatusNotFound)
					App.Logger.Errorln("Error in CounterStorage: ", err)
					return
				} else if retryerr.CheckErrorType(err) {
					if i == 0 {
						time.Sleep(1 * time.Second)
					} else {
						time.Sleep(time.Duration(i+i+1) * time.Second)
					}
				}

			}

			builder := strings.Builder{}
			builder.WriteString(strconv.FormatInt(metricValue, 10))
			metricRes = builder.String()
		} else if metricType == "gauge" {
			var metricValue float64
			for i := 0; i <= 3; i++ {
				metricValue, err = App.Storage.GetGaugeValueByName(metricName)
				if err == nil {
					break
				}
				if !(retryerr.CheckErrorType(err)) {
					http.Error(rw, fmt.Sprintf("Error 404: %s", err), http.StatusNotFound)
					App.Logger.Errorln("Error in GaugeStorage: ", err)
					return
				}

				if i == 0 {
					time.Sleep(1 * time.Second)
				} else {
					time.Sleep(time.Duration(i+i+1) * time.Second)
				}
			}
			metricRes = strconv.FormatFloat(metricValue, 'f', -1, 64)
		} else {
			http.Error(rw, fmt.Sprintf("Error 400: Invalid metric type: %s", metricType), http.StatusBadRequest)
			App.Logger.Errorln("Invalid metric type:", metricType)
			return
		}

		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if App.SecretKey != "" {
			h := hmac.New(sha256.New, []byte(App.SecretKey))
			h.Write([]byte(metricRes))
			signCheck := h.Sum(nil)
			rw.Header().Set("HashSHA256", hex.EncodeToString(signCheck))
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(metricRes))

	}
}

// CheckStorageConnection - handler, that checks if connection to PosthreSQL is alive.
// The function accepts /ping request and return 200, if everything is ok.
func (App *Application) CheckStorageConnection() http.HandlerFunc {
	checkStorageConnectionfunc := func(rw http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

		defer cancel()
		storageAvailable := App.Storage.CheckConnection(ctx)
		if storageAvailable != nil {
			http.Error(rw, storageAvailable.Error(), http.StatusInternalServerError)
			App.Logger.Errorln("Error during Storage connection:", storageAvailable)
		}
		switch ctx.Err() {
		case context.Canceled:
			http.Error(rw, storageAvailable.Error(), http.StatusInternalServerError)
			App.Logger.Errorln("Error during Storage connection:", storageAvailable)
		default:
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
		}
	}
	return http.HandlerFunc(checkStorageConnectionfunc)
}

// GetMetric - handler, that retrieve metrics value from PostgreSQL or in-memory storage and return the value.
// The function gets all data about metrics from request body.
func (App *Application) GetMetric() http.HandlerFunc {
	getMetricfunc := func(rw http.ResponseWriter, r *http.Request) {
		metricData := data.Metrics{}

		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			App.Logger.Errorln("Bad request catched")
			return
		}
		if err = json.Unmarshal(buf.Bytes(), &metricData); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			App.Logger.Errorln("Error during deserialization: ", err)
			return
		}

		if metricData.ID == "" {
			http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
			App.Logger.Errorln("Metric name was not found")
			return
		}
		if metricData.MType == "counter" {
			var metricValue int64
			for i := 0; i <= 3; i++ {
				metricValue, err = App.Storage.GetCounterValueByName(metricData.ID)
				if err == nil {
					break
				}

				if !(retryerr.CheckErrorType(err)) || (i == 3) {
					http.Error(rw, fmt.Sprintf("Error 404: %s", err), http.StatusNotFound)
					App.Logger.Errorln("Error in CounterStorage:", err)
					return
				}

				if i == 0 {
					time.Sleep(1 * time.Second)
				} else {
					time.Sleep(time.Duration(i+i+1) * time.Second)
				}
			}
			metricData.Delta = &metricValue
		} else if metricData.MType == "gauge" {
			var metricValue float64
			for i := 0; i <= 3; i++ {
				metricValue, err = App.Storage.GetGaugeValueByName(metricData.ID)
				if err == nil {
					break
				}
				if !(retryerr.CheckErrorType(err)) || (i == 3) {
					http.Error(rw, fmt.Sprintf("Error 404: %s", err), http.StatusNotFound)
					App.Logger.Errorln("Error in GaugeStorage:", err)
					return
				}

				if i == 0 {
					time.Sleep(1 * time.Second)
				} else {
					time.Sleep(time.Duration(i+i+1) * time.Second)
				}
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

		if App.SecretKey != "" {
			h := hmac.New(sha256.New, []byte(App.SecretKey))
			h.Write(metricDataBytes)
			signCheck := h.Sum(nil)
			rw.Header().Set("HashSHA256", hex.EncodeToString(signCheck))
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write(metricDataBytes)

	}
	return http.HandlerFunc(getMetricfunc)
}

// UpdateAllValues - handler, that updates all values. The function works with pool of data.
func (App *Application) UpdateAllValues() http.HandlerFunc {
	updateAllValuesfunc := func(rw http.ResponseWriter, r *http.Request) {
		metricDataList := make([]data.Metrics, 100)
		var buf bytes.Buffer
		var result []byte

		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		if strings.Contains(r.Header.Get("X-Encrypted"), "rsa") {
			result, err = data.DecryptData(App.CryptoKey, buf.Bytes())
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				App.Logger.Errorln("Error while decrypting data:", err)
				return
			}
		} else {
			result = buf.Bytes()
		}

		reader := bytes.NewReader(result)
		buf.Reset()
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(reader)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				App.Logger.Errorln("Error during unpacking the request: ", err)
				return
			}
			defer gz.Close()

			_, err = buf.ReadFrom(gz)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

		}

		if err := json.Unmarshal(buf.Bytes(), &metricDataList); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			App.Logger.Errorln("Error during deserialization:", err)
			return
		}

		for _, metric := range metricDataList {
			if metric.ID == "" {
				http.Error(rw, "Error 404: Metric name was not found", http.StatusNotFound)
				App.Logger.Errorln("Metric name was not found")
				return
			}
			if (metric.MType != "counter") && (metric.MType != "gauge") {
				http.Error(rw, fmt.Sprintf("Error 400: Metric with name %s invalid metric type : %s", metric.ID, metric.MType), http.StatusBadRequest)
				App.Logger.Errorln(fmt.Sprintf("Metric with name %s invalid metric type : %s", metric.ID, metric.MType))
				return
			}
		}

		for i := 0; i <= 3; i++ {
			err := App.Storage.RepositoryAddAllValues(metricDataList)
			if err == nil {
				break
			}
			if !(retryerr.CheckErrorType(err)) || (i == 3) {
				http.Error(rw, fmt.Sprintf("Error while adding all metrics to storage: %s", err), http.StatusInternalServerError)
				App.Logger.Errorln("Error while adding all metrics to storage", err)
				return
			}

			if i == 0 {
				time.Sleep(1 * time.Second)
			} else {
				time.Sleep(time.Duration(i+i+1) * time.Second)
			}
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
	}

	return http.HandlerFunc(updateAllValuesfunc)
}
