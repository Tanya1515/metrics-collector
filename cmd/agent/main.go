// Agent is used to collect metrics from runtime package,
// such as "HeapAlloc", CPUUtilization and etc, and send
// collected metrics to server.
// Data is collected every 2 seconds (default meaning).
// Data is sent every 10 seconds (default meaning).
package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
	retryerr "github.com/Tanya1515/metrics-collector.git/cmd/errors"
)

var (
	// ReportIntervalFlag - flag for setting up delay for sending metrics to server
	ReportIntervalFlag *int
	// PollIntervalFlag - flag for setting up delay for collecting metrics
	PollIntervalFlag *int
	// serverAddressFlag - flag for setting up server address for sending metrics
	serverAddressFlag *string
	// secretKeyFlag - flag for secret key for
	secretKeyFlag           *string
	cryptoKeyPathFlag       *string
	configFilePathFlag      *string
	limitServerRequestsFlag *int
	buildVersion            string = "N/A"
	buildDate               string = "N/A"
	buildCommit             string = "N/A"
)

func init() {
	ReportIntervalFlag = flag.Int("r", 10, "time duration for sending metrics")
	PollIntervalFlag = flag.Int("p", 2, "time duration for getting metrics")
	serverAddressFlag = flag.String("a", "localhost:8080", "server address")
	secretKeyFlag = flag.String("k", "", "secret key for creating hash")
	limitServerRequestsFlag = flag.Int("l", 1, "limit of requests to server")
	cryptoKeyPathFlag = flag.String("crypto-key", "", "path to key for asymmetrical encryption")
	configFilePathFlag = flag.String("config", "", "path to config file for the application")
}

// MakeMetrics - make list of data.Metrics from map.
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

// GetMetrics - function, that collects all metrics from runtime library
func GetMetrics(ctx context.Context, chanSend chan int64, chanMetrics chan []data.Metrics, timer time.Duration) {
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
		case <-ctx.Done():
			return
		default:
			time.Sleep(timer * time.Second)
		}
		pollCount += 1
	}
}

// GetMetricsUtil - function, that collects total/free memory and utilizatin of every cpu.
func GetMetricsUtil(ctx context.Context, chanSend chan int64, chanMetrics chan []data.Metrics, timer time.Duration) {
	var memStats mem.VirtualMemoryStat
	mapMetrics := make(map[string]float64)
	var pollCount int64
	for {
		freeMemory := memStats.Free
		totalMemory := memStats.Total

		CPUutilization1, _ := cpu.Percent(0, true)

		mapMetrics["TotalMemory"] = float64(totalMemory)
		mapMetrics["FreeMemory"] = float64(freeMemory)
		for key, value := range CPUutilization1 {
			cpuNum := strconv.Itoa(key)
			metricName := "CPUutilization" + cpuNum
			mapMetrics[metricName] = value
		}

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
		case <-ctx.Done():
			return
		default:
			time.Sleep(timer * time.Second)
		}
		pollCount += 1
	}

}

// MakeString - function, that makes request-string for sending metrics to server.
func MakeString(serverAddress string) string {
	builder := strings.Builder{}
	builder.WriteString("http://")
	builder.WriteString(serverAddress)
	builder.WriteString("/updates/")

	return builder.String()
}

func main() {
	fmt.Println("Build version: ", buildVersion)
	fmt.Println("Build date: ", buildDate)
	fmt.Println("Build commit: ", buildCommit)
	var err error
	var cryptoKey []byte
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

	configFilePath, envExists := os.LookupEnv("CONFIG")
	if !(envExists) {
		configFilePath = *configFilePathFlag
	}
	configAgent := new(data.ConfigAgent)
	if configFilePath != "" {

		config, err := os.ReadFile(configFilePath)
		if err != nil {
			fmt.Println("Error while reading config file for agent: ", err)
		}

		err = json.Unmarshal(config, configAgent)
		if err != nil {
			fmt.Println("Error while unmarshaling data from config file for agent: ", err)
		}
	}

	serverAddress, addressExists := os.LookupEnv("ADDRESS")
	if !(addressExists) {
		serverAddress = *serverAddressFlag
	}

	if serverAddress == "localhost:8080" && configFilePath != "" {
		serverAddress = configAgent.ServerAddress
	}

	reportInt := *ReportIntervalFlag
	reportInterval, reportIntExists := os.LookupEnv("REPORT_INTERVAL")
	if reportIntExists {
		reportInt, err = strconv.Atoi(reportInterval)
		if err != nil {
			Logger.Errorln("Error while transforming to int: ", err)
		}
	}

	if reportInt == 10 && configFilePath != "" {
		if configAgent.ReportInterval != "" {
			reportInt, err = strconv.Atoi(strings.Split(configAgent.ReportInterval, "s")[0])
			if err != nil {
				Logger.Errorln("Error while report interval transforming to int: ", err)
			}
		}
	}

	cryptoKeyPath, envExists := os.LookupEnv("CRYPTO_KEY")
	if !(envExists) {
		cryptoKeyPath = *cryptoKeyPathFlag
	}

	if cryptoKeyPath == "" && configFilePath != "" {
		cryptoKeyPath = configAgent.CryptoKeyPath
	}

	if cryptoKeyPath != "" {
		cryptoKey, err = os.ReadFile(cryptoKeyPath)
		if err != nil {
			fmt.Println("Error while reading file with crypto key: ", err)
		}
	}

	pollInt := *PollIntervalFlag
	pollInterval, pollIntervalExist := os.LookupEnv("POLL_INTERVAL")
	if pollIntervalExist {
		pollInt, err = strconv.Atoi(pollInterval)
		if err != nil {
			Logger.Errorln("Error while transforming to int: ", err)
		}
	}

	if pollInt == 2 && configFilePath != "" {
		if configAgent.PollInterval != "" {
			pollInt, err = strconv.Atoi(strings.Split(configAgent.PollInterval, "s")[0])
			if err != nil {
				Logger.Errorln("Error while poll interval transforming to int: ", err)
			}
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

	if limitRequests == 1 && configFilePath != "" {
		limitRequests = configAgent.LimitServerRequests
	}

	secretKeyHash, secretKeyExists := os.LookupEnv("KEY")
	if !(secretKeyExists) {
		secretKeyHash = *secretKeyFlag
	}

	if secretKeyHash == "" && configFilePath != "" {
		secretKeyHash = configAgent.SecretKey
	}
	requestString := MakeString(serverAddress)

	gracefulSutdown := make(chan os.Signal, 1)
	shutdown := make(chan struct{})
	signal.Notify(gracefulSutdown, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go GetMetrics(ctx, chansPollCount[0], chanMetrics, time.Duration(pollInt))

	go GetMetricsUtil(ctx, chansPollCount[1], chanMetrics, time.Duration(pollInt))
	var wg sync.WaitGroup
	go func() {
		<-gracefulSutdown
		close(shutdown)
		Logger.Infoln("Wait for sending all metrics to server")
		wg.Wait()
		close(resultChannel)
	}()

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
			select {
			case <-shutdown:
				for value := range resultChannel {
					if value != nil {
						Logger.Errorln("Error while sending metrics: ", value)
					} else {
						Logger.Infoln("Metrics were sent successfully")
					}
				}
				Logger.Infoln("Wait for canceling goroutines, that gather metrics")
				cancel()
				Logger.Infoln("Stop agent")
				return
			default:
				for i := range chansPollCount {
					chansPollCount[i] <- -1
					chanPollCount := chansPollCount[i]
					wg.Add(1)
					go func() {
						Logger.Infoln("Start sending metrics to server")
						defer wg.Done()
						metrics := <-chanMetrics
						var sign []byte

						compressedMetrics, err := data.Compress(&metrics)
						if err != nil {
							resultChannel <- err
							return

						}
						if cryptoKey != nil {
							compressedMetrics, err = data.EncryptData(compressedMetrics, cryptoKey)
							if err != nil {
								resultChannel <- err
								return
							}
						}
						if secretKeyHash != "" {
							h := hmac.New(sha256.New, []byte(secretKeyHash))
							h.Write(compressedMetrics)
							sign = h.Sum(nil)
						}

						sem <- struct{}{}
						defer func() { <-sem }()
						for i := 0; i <= 3; i++ {
							if secretKeyHash != "" && cryptoKey != nil {
								_, err = client.R().
									SetHeader("Content-Type", "application/json").
									SetHeader("Content-Encoding", "gzip").
									SetHeader("X-Encrypted", "rsa").
									SetHeader("HashSHA256", hex.EncodeToString(sign)).
									SetBody(compressedMetrics).
									Post(requestString)
							} else if cryptoKey != nil {
								_, err = client.R().
									SetHeader("Content-Type", "application/json").
									SetHeader("Content-Encoding", "gzip").
									SetHeader("X-Encrypted", "rsa").
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
								break
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
}
