package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
	retryerr "github.com/Tanya1515/metrics-collector.git/cmd/errors"
)

var (
	reportIntervalFlag      *int
	pollIntervalFlag        *int
	serverAddressFlag       *string
	secretKeyFlag           *string
	limitServerRequestsFlag *int
)

func init() {
	reportIntervalFlag = flag.Int("r", 0, "time duration for sending metrics")
	pollIntervalFlag = flag.Int("p", 2, "time duration for getting metrics")
	serverAddressFlag = flag.String("a", "localhost:8080", "server address")
	secretKeyFlag = flag.String("k", "", "secret key for creating hash")
	limitServerRequestsFlag = flag.Int("l", 1, "limit of requests to server")
}

func MakeMetrics(mapMetrics map[string]float64, pollCount int64) []data.Metrics {
	metrics := make([]data.Metrics, len(mapMetrics)+1)
	i := 0

	for metricName, metricValue := range mapMetrics {
		metricData := data.Metrics{}
		metricData.ID = metricName
		metricData.MType = "gauge"
		metricData.Value = &metricValue
		metrics[i] = metricData
		i += 1
	}
	metricData := data.Metrics{}
	metricData.ID = "PollCount"
	metricData.MType = "counter"
	metricData.Delta = &pollCount
	metrics[i] = metricData

	return metrics
}

// Alternative variant of structure processing: variable := float64(memStats.Alloc)
func GetMetrics(chanSend chan int64, chanMetrics chan []data.Metrics, timer time.Duration) {
	var memStats runtime.MemStats
	mapMetrics := make(map[string]float64)
	var pollCount int64
	for {
		runtime.ReadMemStats(&memStats)
		mapMetrics["Alloc"] = float64(memStats.Alloc)
		mapMetrics["BuckHashSys"] = float64(memStats.BuckHashSys)
		mapMetrics["Frees"] = float64(memStats.Frees)
		mapMetrics["GCCPUFraction"] = float64(memStats.GCCPUFraction)
		mapMetrics["GCSys"] = float64(memStats.GCSys)
		mapMetrics["HeapAlloc"] = float64(memStats.HeapAlloc)
		mapMetrics["HeapIdle"] = float64(memStats.HeapIdle)
		mapMetrics["HeapInuse"] = float64(memStats.HeapInuse)
		mapMetrics["HeapObjects"] = float64(memStats.HeapObjects)
		mapMetrics["HeapReleased"] = float64(memStats.HeapReleased)
		mapMetrics["HeapSys"] = float64(memStats.HeapSys)
		mapMetrics["LastGC"] = float64(memStats.LastGC)
		mapMetrics["Lookups"] = float64(memStats.Lookups)
		mapMetrics["MCacheInuse"] = float64(memStats.MCacheInuse)
		mapMetrics["MCacheSys"] = float64(memStats.MCacheSys)
		mapMetrics["MSpanInuse"] = float64(memStats.MSpanInuse)
		mapMetrics["MSpanSys"] = float64(memStats.MSpanSys)
		mapMetrics["Mallocs"] = float64(memStats.Mallocs)
		mapMetrics["NextGC"] = float64(memStats.NextGC)
		mapMetrics["NumForcedGC"] = float64(memStats.NumForcedGC)
		mapMetrics["NumGC"] = float64(memStats.NumGC)
		mapMetrics["OtherSys"] = float64(memStats.OtherSys)
		mapMetrics["PauseTotalNs"] = float64(memStats.PauseTotalNs)
		mapMetrics["StackInuse"] = float64(memStats.StackInuse)
		mapMetrics["StackSys"] = float64(memStats.StackSys)
		mapMetrics["Sys"] = float64(memStats.Sys)
		mapMetrics["TotalAlloc"] = float64(memStats.TotalAlloc)
		mapMetrics["RandomValue"] = rand.Float64()

		select {
		case signal, ok := <-chanSend:
			if !ok {
				return
			}
			if signal == -1 {
				metrics := MakeMetrics(mapMetrics, pollCount)
				chanMetrics <- metrics
			} else {
				pollCount = signal
			}
		default:
			time.Sleep(timer * time.Second)
		}
		pollCount += 1
	}
}

func GetMetricsUtil(chanSend chan int64, chanMetrics chan []data.Metrics, timer time.Duration) {
	var memStats mem.VirtualMemoryStat
	mapMetrics := make(map[string]float64)
	var pollCount int64
	for {
		freeMemory := memStats.Free
		totalMemory := memStats.Total

		CPUutilization1, _ := cpu.Percent(0, true)

		mapMetrics["TotalMemory"] = float64(totalMemory)
		mapMetrics["FreeMemory"] = float64(freeMemory)
		mapMetrics["CPUutilization1"] = CPUutilization1[0]
		select {
		case signal, ok := <-chanSend:
			if !ok {
				return
			}
			if signal == -1 {
				metrics := MakeMetrics(mapMetrics, pollCount)
				chanMetrics <- metrics
			} else {
				pollCount = signal
			}
		default:
			time.Sleep(timer * time.Second)
		}
		pollCount += 1
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
	chansPollCount := []chan int64{
		make(chan int64),
		make(chan int64),
	}
	resultChannel := make(chan error)
	defer close(chansPollCount[0])
	defer close(chansPollCount[1])
	chanMetrics := make(chan []data.Metrics, 10)

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

	limitRequests := *limitServerRequestsFlag
	limitReq, limitReqExist := os.LookupEnv("RATE_LIMIT")
	if limitReqExist {
		limitRequests, err = strconv.Atoi(limitReq)
		if err != nil {
			Logger.Errorln("Error while transforming to int: ", err)
		}
	}

	secretKeyHash, secretKeyExists := os.LookupEnv("KEY")
	if !(secretKeyExists) {
		secretKeyHash = *secretKeyFlag
	}

	requestString := MakeString(serverAddress)
	go GetMetrics(chansPollCount[0], chanMetrics, time.Duration(pollInt))
	go GetMetricsUtil(chansPollCount[1], chanMetrics, time.Duration(pollInt))

	sem := make(chan struct{}, limitRequests)

	for {
		select {
		case result := <-resultChannel:
			err = result
			if err != nil {
				Logger.Errorln("Error while sending metrics: ", err)
			} else {
				Logger.Infoln("Metrics were sent successfully")
			}
		default:
			Logger.Infoln("Agent begin sleeping")
			time.Sleep(time.Duration(reportInt) * time.Second)

			for i := range chansPollCount {
				Logger.Infoln("Start sending metrics to server")
				chansPollCount[i] <- -1
				chanPollCount := chansPollCount[i]

				go func() {
					metrics := <-chanMetrics
					var sign []byte

					compressedMetrics, err := data.Compress(&metrics)
					if err != nil {
						resultChannel <- err
						return

					}
					if secretKeyHash != "" {
						h := hmac.New(sha256.New, []byte(secretKeyHash))
						h.Write(compressedMetrics)
						sign = h.Sum(nil)
					}

					sem <- struct{}{}
					defer func() { <-sem }()
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
							chanPollCount <- 0
						}
						if !(retryerr.CheckErrorType(err)) || (i == 3) {
							break
						}

						if i == 0 {
							time.Sleep(1 * time.Second)
						} else {
							time.Sleep(time.Duration(i+i+1) * time.Second)
						}

					}
					resultChannel <- err
				}()
			}
		}
	}
}
