package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	reportIntervalFlag *int
	pollIntervalFlag   *int
	serverAddressFlag  *string
)

func init() {
	reportIntervalFlag = flag.Int("r", 10, "time duration for sending metrics")
	pollIntervalFlag = flag.Int("p", 2, "time duration for getting metrics")
	serverAddressFlag = flag.String("a", "localhost:8080", "server address")
}


type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (metricData *Metrics) Compress() ([]byte, error) {
	var b bytes.Buffer

	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed init compress writer: %v", err)
	}

	metricDataBytes, err := json.Marshal(metricData)
	if err != nil {
		return nil, fmt.Errorf("failed to compress data: %v", err)
	}
	_, err = w.Write(metricDataBytes)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}

	return b.Bytes(), nil
}

func CheckValue(fieldName string) bool {
	gaugeMetrics := [...]string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"}

	for _, valueMetric := range gaugeMetrics {
		if valueMetric == fieldName {
			return true
		}
	}
	return false
}

func GetMetrics(mapMetrics *map[string]interface{}, PollCount *int64, timer time.Duration, mutex *sync.RWMutex) {
	var memStats runtime.MemStats
	for {
		runtime.ReadMemStats(&memStats)
		val := reflect.ValueOf(memStats)

		mutex.Lock()
		for fieldIndex := 0; fieldIndex < val.NumField(); fieldIndex++ {
			field := val.Field(fieldIndex)
			fieldName := val.Type().Field(fieldIndex).Name
			if CheckValue(fieldName) {
				(*mapMetrics)[fieldName] = field
			}
		}
		(*mapMetrics)["RandomValue"] = rand.Float64()
		(*PollCount) += 1
		mutex.Unlock()
		time.Sleep(timer * time.Second)
	}
}

func MakeString(serverAddress string) string {
	builder := strings.Builder{}
	builder.WriteString("http://")
	builder.WriteString(serverAddress)
	builder.WriteString("/update")

	return builder.String()
}

func main() {
	var err error
	fmt.Println("Start agent for collecting metrics")
	mapMetrics := make(map[string]interface{}, 20)
	var PollCount int64
	var mutex sync.RWMutex

	client := resty.New()

	flag.Parse()

	serverAddress, addressExists := os.LookupEnv("ADDRESS")
	if !(addressExists) {
		serverAddress = *serverAddressFlag
	}

	reportInt := *reportIntervalFlag
	reportInterval, reportIntExists := os.LookupEnv("REPORT_INTERVAL")
	if reportIntExists {
		reportInt, err = strconv.Atoi(reportInterval)
		if err != nil {
			fmt.Printf("Error while transforming to int: %s\n", err)
		}
	}

	pollInt := *pollIntervalFlag
	pollInterval, pollIntervalExist := os.LookupEnv("POLL_INTERVAL")
	if pollIntervalExist {
		pollInt, err = strconv.Atoi(pollInterval)
		if err != nil {
			fmt.Printf("Error while transforming to int: %s\n", err)
		}
	}

	metricData := Metrics{}
	requestString := MakeString(serverAddress)
	var metricValueF64 float64
	go GetMetrics(&mapMetrics, &PollCount, time.Duration(pollInt), &mutex)

	for {
		time.Sleep(time.Duration(reportInt) * time.Second)
		mutex.RLock()
		for metricName, metricValue := range mapMetrics {
			metricData.ID = metricName
			metricData.MType = "gauge"
			metricValueF64, err = strconv.ParseFloat(fmt.Sprint(metricValue), 64)
			if err != nil {
				fmt.Printf("Error while parsing metric %s: %s", metricName, err)
			}
			metricData.Value = &metricValueF64
			compressedMetric, err := metricData.Compress()
			if err != nil {
				fmt.Println(err)
			}
			_, err = client.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("Content-Encoding", "gzip").
				SetHeader("Accept-Encoding", "gzip").
				SetBody(compressedMetric).
				Post(requestString)
			if err != nil {
				fmt.Printf("Error while sending metric %s: %s\n", metricName, err)
			}
		}
		metricData.ID = "PollCount"
		metricData.MType = "counter"
		metricData.Delta = &PollCount
		_, err = client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(metricData).
			Post(requestString)
		if err != nil {
			fmt.Printf("Error while sending PollCounter for metric PollCount: %s", err)
		} else {
			PollCount = 0
		}

		mutex.RUnlock()
	}
}
