package main

import (
	"flag"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	reportInterval *time.Duration
	pollInterval   *time.Duration
	serverAddress  *string
)

func init() {
	reportInterval = flag.Duration("r", 10, "time duration for sending metrics")
	pollInterval = flag.Duration("p", 2, "time duration for getting metrics")
	serverAddress = flag.String("a", "localhost:8080", "server address")
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

func GetMetrics(mapMetrics *map[string]interface{}, PollCount *int, timer time.Duration, mutex *sync.RWMutex) {
	var memStats runtime.MemStats
	for {
		fmt.Println(timer)
		time.Sleep(timer * time.Second)
		fmt.Println("Hello!")
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
	}
}

func MakeString(serverAddress, metricName, metricValue, metricType string) string {
	builder := strings.Builder{}

	if metricType == "gauge" {
		builder.WriteString("http://")
		builder.WriteString(serverAddress)
		builder.WriteString("/update/gauge/")
		builder.WriteString(metricName)
		builder.WriteString("/")
		builder.WriteString(metricValue)
	}

	if metricType == "counter" {
		builder.WriteString("http://")
		builder.WriteString(serverAddress)
		builder.WriteString("/update/counter/PollCount")
		builder.WriteString(metricName)
		builder.WriteString("/")
		builder.WriteString(metricValue)
	}

	return builder.String()
}

func main() {

	fmt.Println("Start agent")
	mapMetrics := make(map[string]interface{}, 20)
	PollCount := 0
	client := resty.New()

	var mutex sync.RWMutex

	flag.Parse()

	go GetMetrics(&mapMetrics, &PollCount, (*pollInterval), &mutex)

	for {
		time.Sleep((*reportInterval) * time.Second)
		mutex.RLock()
		for metricName, metricValue := range mapMetrics {
			fmt.Println("Send metrics")
			metricValueStr := fmt.Sprint(metricValue)
			requestString := MakeString(*serverAddress, metricName, metricValueStr, "gauge")
			_, err := client.R().
				SetHeader("Content-Type", "text/plain").
				Post(requestString)
			if err != nil {
				fmt.Printf("Error while sending metric %s: %s", metricName, err)
			}

			requestString = MakeString(*serverAddress, metricName, fmt.Sprint(PollCount), "counter")
			_, err = client.R().
				SetHeader("Content-Type", "text/plain").
				Post(requestString)
			if err != nil {
				fmt.Printf("Error while sending PollCounter for metric %s: %s", metricName, err)
			}
		}
		mutex.RUnlock()
	}
}
