package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

	"go.uber.org/zap"

	"github.com/go-resty/resty/v2"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
	retryerr "github.com/Tanya1515/metrics-collector.git/cmd/errors"
)

var (
	reportIntervalFlag *int
	pollIntervalFlag   *int
	serverAddressFlag  *string
	secretKeyFlag      *string
)

func init() {
	reportIntervalFlag = flag.Int("r", 10, "time duration for sending metrics")
	pollIntervalFlag = flag.Int("p", 2, "time duration for getting metrics")
	serverAddressFlag = flag.String("a", "localhost:8080", "server address")
	secretKeyFlag = flag.String("k", "", "secret key for creating hash")
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

// Alternative variant of structure processing: variable := float64(memStats.Alloc)
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
	builder.WriteString("/updates/")

	return builder.String()
}

func main() {
	var err error
	mapMetrics := make(map[string]interface{}, 20)
	var PollCount int64
	var mutex sync.RWMutex

	client := resty.New()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	Logger := *logger.Sugar()

	Logger.Infow(
		"Starting agent for collecting metrics",
	)

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
			Logger.Errorln("Error while transforming to int: ", err)
		}
	}

	pollInt := *pollIntervalFlag
	pollInterval, pollIntervalExist := os.LookupEnv("POLL_INTERVAL")
	if pollIntervalExist {
		pollInt, err = strconv.Atoi(pollInterval)
		if err != nil {
			Logger.Errorln("Error while transforming to int: ", err)
		}
	}

	secretKeyHash, secretKeyExists := os.LookupEnv("KEY")
	if !(secretKeyExists) {
		secretKeyHash = *secretKeyFlag
	}

	requestString := MakeString(serverAddress)
	go GetMetrics(&mapMetrics, &PollCount, time.Duration(pollInt), &mutex)

	for {
		time.Sleep(time.Duration(reportInt) * time.Second)
		mutex.RLock()
		i := 0
		metrics := make([]data.Metrics, 29)
		for metricName, metricValue := range mapMetrics {
			metricData := data.Metrics{}
			metricData.ID = metricName
			metricData.MType = "gauge"
			metricValueF64, err := strconv.ParseFloat(fmt.Sprint(metricValue), 64)
			if err != nil {
				Logger.Errorln("Error while parsing metric ", metricName, ": ", err)
			}
			metricData.Value = &metricValueF64
			metrics[i] = metricData
			i += 1
		}
		metricData := data.Metrics{}
		metricData.ID = "PollCount"
		metricData.MType = "counter"
		metricData.Delta = &PollCount
		metrics[i] = metricData

		var sign []byte

		compressedMetrics, err := data.Compress(&metrics)
		if err != nil {
			Logger.Errorln("Error while compressing data: ", err)
		}
		if secretKeyHash != "" {
			h := hmac.New(sha256.New, []byte(secretKeyHash))
			h.Write(compressedMetrics)
			sign = h.Sum(nil)
		}
		Logger.Infoln("Send metrics to server")

		for i := 0; i <= 3; i++ {
			if secretKeyHash != "" {
				_, err = client.R().
					SetHeader("Content-Type", "application/json").
					SetHeader("Content-Encoding", "gzip").
					SetHeader("HashSHA256", hex.EncodeToString(sign)).
					SetBody(compressedMetrics).
					Post(requestString)
			} else {
				_, err = client.R().
					SetHeader("Content-Type", "application/json").
					SetHeader("Content-Encoding", "gzip").
					SetBody(compressedMetrics).
					Post(requestString)
			}
			if err == nil {
				PollCount = 0
				break
			}
			if !(retryerr.CheckErrorType(err)) || (i == 3) {
				Logger.Errorln("Error while sending metrics: ", err)
				break
			}

			Logger.Errorln("Network error:", err, ", retry: ", i)
			if i == 0 {
				time.Sleep(1 * time.Second)
			} else {
				time.Sleep(time.Duration(i+i+1) * time.Second)
			}
		}

		mutex.RUnlock()
	}
}
