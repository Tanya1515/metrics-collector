package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
	retryerr "github.com/Tanya1515/metrics-collector.git/cmd/errors"
	pb "github.com/Tanya1515/metrics-collector.git/cmd/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *MetricsServer) SendMetrics(inOut grpc.ClientStreamingServer[pb.Metrics, pb.MetricsResponse]) error {

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

func (server *MetricsServer) GetAllMetrics(ctx context.Context, in *emptypb.Empty) (*pb.Metrics, error) {
	var err error
	var result pb.Metrics
	allMetrics := make([]*pb.Metric, 100)
	var i int

	counterMetrics := make(map[string]int64, 100)
	gaugeMetrics := make(map[string]float64, 100)

	counterMetrics, err = server.App.Storage.GetAllCounterMetrics()
	if err != nil {
		server.App.Logger.Errorln("Error, while getting all counter metrics: ", err)
		return nil, fmt.Errorf("error, while getting all counter metrics: %w", err)
	}

	gaugeMetrics, err = server.App.Storage.GetAllGaugeMetrics()
	if err != nil {
		server.App.Logger.Errorln("Error, while getting all gauge metrics: ", err)
		return nil, fmt.Errorf("error, while getting all gauge metrics: %w", err)
	}

	for metricName, metric := range counterMetrics {
		if i == 100 {
			allMetrics = append(allMetrics, &pb.Metric{Id: metricName, Mtype: pb.Metric_COUNTER, MetricValue: &pb.Metric_Delta{Delta: metric}})
		} else {
			allMetrics[i].Id = metricName
			allMetrics[i].Mtype = pb.Metric_COUNTER
			allMetrics[i].MetricValue = &pb.Metric_Delta{Delta: metric}
		}
	}

	for metricName, metric := range gaugeMetrics {
		if i == 100 {
			allMetrics = append(allMetrics, &pb.Metric{Id: metricName, Mtype: pb.Metric_GAUGE, MetricValue: &pb.Metric_Value{Value: metric}})
		} else {
			allMetrics[i].Id = metricName
			allMetrics[i].Mtype = pb.Metric_GAUGE
			allMetrics[i].MetricValue = &pb.Metric_Value{Value: metric}
		}
	}

	result.Metrics = allMetrics

	return &result, err
}

func (server *MetricsServer) GetMetric(ctx context.Context, inMetric *pb.Metric) (*pb.Metric, error) {
	var err error

	if inMetric.Id == "" {
		server.App.Logger.Errorln("Empry metric name recieved")
		return inMetric, errors.New("empry metric name recieved")
	}

	if strings.ToLower(inMetric.Mtype.String()) == "counter" {
		delta, err := server.App.Storage.GetCounterValueByName(inMetric.Id)
		if err != nil {
			server.App.Logger.Errorln("Error while adding counter value: ", err)
			return inMetric, fmt.Errorf("error while adding counter metric: %w", err)
		}
		inMetric.MetricValue = &pb.Metric_Delta{Delta: delta}
	} else if strings.ToLower(inMetric.Mtype.String()) == "gauge" {
		value, err := server.App.Storage.GetGaugeValueByName(inMetric.Id)
		if err != nil {
			server.App.Logger.Errorln("Error while adding gauge value: ", err)
			return inMetric, fmt.Errorf("error while adding gauge metric: ", err)
		}
		inMetric.MetricValue = &pb.Metric_Value{Value: value}
	} else {
		server.App.Logger.Errorln("Incorrect metric type recieved")
		return inMetric, errors.New("incorrect metric type recieved")
	}

	return inMetric, err
}

func (server *MetricsServer) PostMetric(ctx context.Context, inMetric *pb.Metric) (*pb.MetricsResponse, error) {
	var err error

	var response pb.MetricsResponse

	if inMetric.Id == "" {
		response.Error = "Empry metric name recieved"
		server.App.Logger.Errorln("Empry metric name recieved")
		return &response, errors.New("empry metric name recieved")
	}

	if strings.ToLower(inMetric.Mtype.String()) == "counter" {
		err = server.App.Storage.RepositoryAddCounterValue(inMetric.Id, inMetric.GetDelta())
		if err != nil {
			response.Error = "Error while adding counter value"
			server.App.Logger.Errorln("Error while adding counter value: ", err)
			return &response, fmt.Errorf("error while adding counter metric: %w", err)
		}
	} else if strings.ToLower(inMetric.Mtype.String()) == "gauge" {
		err = server.App.Storage.RepositoryAddGaugeValue(inMetric.Id, inMetric.GetValue())
		if err != nil {
			response.Error = "Error while adding gauge value"
			server.App.Logger.Errorln("Error while adding gauge value: ", err)
			return &response, fmt.Errorf("error while adding gauge metric: ", err)
		}
	} else {
		response.Error = "Incorrect metric type recieved"
		server.App.Logger.Errorln("Incorrect metric type recieved")
		return &response, errors.New("incorrect metric type recieved")
	}

	return nil, err
}
