package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
	retryerr "github.com/Tanya1515/metrics-collector.git/cmd/errors"
	pb "github.com/Tanya1515/metrics-collector.git/cmd/grpc/proto"
	"google.golang.org/grpc"
)

func (server *MetricsServer) SendMetrics(inOut grpc.ClientStreamingServer[pb.MetricsRequest, pb.MetricsResponse]) error {

	var response pb.MetricsResponse

	metricsRecieved, err := inOut.Recv()
	if err != nil {
		server.App.Logger.Errorln("Recieved error while processing recieving metrics ", err)
		response.Error = err.Error()
		inOut.SendAndClose(&response)
		return err
	}

	metrics := metricsRecieved.GetMetrics()
	metricDataList := make([]data.Metrics, len(metrics))

	for key, metric := range metrics {
		if metric.Id == "" {
			response.Error = "Empry metric name recieved"
			inOut.SendAndClose(&response)
			server.App.Logger.Errorln("Empry metric name recieved")
			return errors.New("empry metric name recieved")
		}
		metricDataList[key].ID = metric.Id
		if strings.ToLower(metric.Mtype.String()) == "counter" {
			metricDataList[key].MType = "counter"
			counterDelta := metric.GetDelta()
			metricDataList[key].Delta = &counterDelta
		} else if strings.ToLower(metric.Mtype.String()) == "gauge" {
			metricDataList[key].MType = "gauge"
			gaugeValue := metric.GetValue()
			metricDataList[key].Value = &gaugeValue
		} else {
			response.Error = "Incorrect metric type recieved"
			inOut.SendAndClose(&response)
			server.App.Logger.Errorln("Incorrect metric type recieved")
			return errors.New("incorrect metric type recieved")
		}
	}

	for i := 0; i <= 3; i++ {
		err := server.App.Storage.RepositoryAddAllValues(metricDataList)
		if err == nil {
			break
		}
		if !(retryerr.CheckErrorType(err)) || (i == 3) {
			response.Error = fmt.Sprintf("Error while adding all metrics to storage: %s", err)
			inOut.SendAndClose(&response)
			server.App.Logger.Errorln("Error while adding all metrics to storage", err)
			return err
		}

		if i == 0 {
			time.Sleep(1 * time.Second)
		} else {
			time.Sleep(time.Duration(i+i+1) * time.Second)
		}
	}

	return nil
}
