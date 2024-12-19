package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"strings"
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

func MakeString (metricName, metricValue, metricType string) string{
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


func SendMetrics(mapMetrics map[string]interface{}, PollCount int) error {
	for metricName, metricValue := range mapMetrics {
		metricValueStr := fmt.Sprint(metricValue)
		request_string := MakeString(metricName, metricValueStr, "gauge")
		_, err := http.Post(request_string, "text/html", nil)
		if err != nil {
			return err
		}
		request_string = MakeString(metricName, fmt.Sprint(PollCount), "counter")
		_, err = http.Post(request_string, "text/html", nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	mapMetrics := make(map[string]interface{}, 20)
	PollCount := 0
	//timer - ? возможно, придется вкатывать goroutine с таймером
	GetMetrics(&mapMetrics, &PollCount)

	SendMetrics(mapMetrics, PollCount)

}
