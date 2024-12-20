package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

func CheckValue(fieldName string) bool {
	gaugeMetrics := [...]string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"}

	for _, valueMetric := range gaugeMetrics {
		if valueMetric == fieldName {
			return true
		}
	}
	return false
}

func GetMetrics(mapMetrics *map[string]interface{}, PollCount *int) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	val := reflect.ValueOf(memStats)

	for fieldIndex := 0; fieldIndex < val.NumField(); fieldIndex++ {
		field := val.Field(fieldIndex)
		fieldName := val.Type().Field(fieldIndex).Name
		if CheckValue(fieldName) {
			(*mapMetrics)[fieldName] = field
		}
	}
	(*mapMetrics)["RandomValue"] = rand.Float64()
	(*PollCount) += 1
}

func MakeString(metricName, metricValue, metricType string) string {
	builder := strings.Builder{}

	if metricType == "gauge" {
		builder.WriteString("http://127.0.0.1:8080/update/gauge/")
		builder.WriteString(metricName)
		builder.WriteString("/")
		builder.WriteString(metricValue)
	}

	if metricType == "counter" {
		builder.WriteString("http://127.0.0.1:8080/update/counter/PollCount")
		builder.WriteString(metricName)
		builder.WriteString("/")
		builder.WriteString(metricValue)
	}

	return builder.String()
}

func goSendMetrics(PCCh chan int, mapCh chan map[string]interface{}) {
	client := resty.New()

	for {
		PollCount := <-PCCh
		mapMetrics := <-mapCh

		for metricName, metricValue := range mapMetrics {
			metricValueStr := fmt.Sprint(metricValue)
			requestString := MakeString(metricName, metricValueStr, "gauge")
			_, err := client.R().
				SetHeader("Content-Type", "text/plain").
				Post(requestString)
			if err != nil {
				fmt.Printf("Error while sending metric %s: %s", metricName, err)
			}

			requestString = MakeString(metricName, fmt.Sprint(PollCount), "counter")
			_, err = client.R().
				SetHeader("Content-Type", "text/plain").
				Post(requestString)
			if err != nil {
				fmt.Printf("Error while sending PollCounter for metric %s: %s", metricName, err)
			}
		}
	}
}

func main() {

	fmt.Println("Start agent")
	mapMetrics := make(map[string]interface{}, 20)
	PollCount := 0

	pollCountCh := make(chan int)

	mapCh := make(chan map[string]interface{})

	go goSendMetrics(pollCountCh, mapCh)

	for {
		time.Sleep(2 * time.Second)
		GetMetrics(&mapMetrics, &PollCount)

		if PollCount%5 == 0 {
			pollCountCh <- PollCount
			mapCh <- mapMetrics
		}
	}
}
