package main

import (
	"bufio"
	"encoding/json"
	"os"
	"time"
)

func (App *Application) Store() error {

	file, err := os.OpenFile(App.FileStore, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		App.Logger.Errorln("Error while openning file: %s", err)
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
			App.Logger.Errorln("Error while metric deserialization: ", err)
			return err
		}

		if metric.MType == "gauge" {
			err = App.Storage.RepositoryAddGaugeValue(metric.ID, *metric.Value)
			if err != nil {
				App.Logger.Errorln("Error while adding gauge metric %s to repository %s", metric.ID, err)
			}
		}

		if metric.MType == "counter" {
			err = App.Storage.RepositoryAddValue(metric.ID, *metric.Delta)
			if err != nil {
				App.Logger.Errorln("Error while adding counter metric %s to repository %s", metric.ID, err)
			}
		}
	}
}

func (App *Application) SaveMetrics(timer time.Duration) {

	gaugeMetric := Metrics{ID: "", MType: "gauge"}
	counterMetric := Metrics{ID: "", MType: "counter"}
	for {
		App.Logger.Infoln("Write data to backup file")
		file, err := os.OpenFile(App.FileStore, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			App.Logger.Errorln("Error while openning file: %s", err)
		}
		allGaugeMetrics, err := App.Storage.GetAllGaugeMetrics()
		if err != nil {
			App.Logger.Errorln("%s", err)
		}
		for metricName, metricValue := range allGaugeMetrics {
			gaugeMetric.ID = metricName
			gaugeMetric.Value = &metricValue

			metricBytes, err := json.Marshal(gaugeMetric)
			if err != nil {
				App.Logger.Errorln("Error while marshalling GaugeMetric: %s", err)
			}
			_, err = file.Write(metricBytes)
			if err != nil {
				App.Logger.Errorln("Error while writing metric info to file: %s", err)
			}
			_, err = file.WriteString("\n")
			if err != nil {
				App.Logger.Errorln("Error while writting line transition: %s", err)
			}
		}
		allCounterMetrics, err := App.Storage.GetAllCounterMetrics()
		if err != nil {
			App.Logger.Errorln("%s", err)
		}
		for metricName, metricValue := range allCounterMetrics {
			counterMetric.ID = metricName
			counterMetric.Delta = &metricValue

			metricBytes, err := json.Marshal(counterMetric)
			if err != nil {
				App.Logger.Errorln("Error while marshalling CounterMetric: %s", err)
			}
			_, err = file.Write(metricBytes)
			if err != nil {
				App.Logger.Errorln("Error while writing metric info to file: %s", err)
			}
			_, err = file.WriteString("\n")
			if err != nil {
				App.Logger.Errorln("Error while writting line transition: %s", err)
			}
		}
		err = file.Close()
		if err != nil {
			App.Logger.Errorln("Error while closing file: %s", err)
		}

		time.Sleep(timer * time.Second)
	}

}
