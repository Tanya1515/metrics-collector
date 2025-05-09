// Data contains basic types and their methods of the project.
package data

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"

	pb "github.com/Tanya1515/metrics-collector.git/cmd/grpc/proto"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

// ResultMetrics - type, that contains all gauge and counter metrics for representation them in HTML-format.
type ResultMetrics struct {
	// GaugeMetrics contains all gauge metrics.
	GaugeMetrics string
	// CounterMetrics contains all counter metrics.
	CounterMetrics string
}

// Metrics - type, that describes all fields of recieved/saved metrics.
type Metrics struct {
	// ID - name of the metric
	ID string `json:"id"`
	// MType - type of the metric, available values are counter and gauge
	MType string `json:"type"`
	// Delta - int64 value, that is specified for counter metric
	Delta *int64 `json:"delta,omitempty"`
	// Value - float64 value, that is specified for gauge metric
	Value *float64 `json:"value,omitempty"`
}

// ConfigApp - type, that describes all fields of the application config file
type ConfigApp struct {
	// ServerAddress - server address
	ServerAddress string `json:"address"`
	// StoreInterval - time duration for saving metrics
	StoreInterval string `json:"store_interval"`
	// FileStorePath - filename for storing metrics
	FileStorePath string `json:"store_file"`
	// Restore - flag for storing all info
	Restore bool `json:"restore"`
	// PostgreSQL - credentials for database
	PostgreSQL string `json:"database_dsn"`
	// SecretKey - secret key for hashing data
	SecretKey string `json:"secret_key"`
	// CryptoKeyPath - path to key for asymmetrical encryption
	CryptoKeyPath string `json:"crypto_key"`
	// TrustedSubnet - CIDR, for detecting if agent IP is trusted
	TrustedSubnet string `json:"trusted_subnet"`
	// GRPC - use grpc protocol for getting metrics from agent
	GRPC bool `json:"grpc"`
}

// ConfigAgent - type, that describes all fields of the agent configuration
type ConfigAgent struct {
	// ReportInterval - time duration for sending metrics
	ReportInterval string `json:"report_interval"`
	// PollInterval - time duration for getting metrics
	PollInterval string `json:"poll_interval"`
	// ServerAddress - server address for sending metrics
	ServerAddress string `json:"address"`
	// SecretKey - secret key for creating hash
	SecretKey string `json:"secret_key"`
	// CryptoKeyPath - limit of requests to server
	CryptoKeyPath string `json:"crypto_key"`
	// LimitServerRequests - path to key for asymmetrical encryption
	LimitServerRequests int `json:"limit_requests"`
	// GRPC - use grpc protocol for sending metrics to server
	GRPC bool `json:"grpc"`
}

// Compress - function for compressing list of metrics to slice of bytes
func Compress(metricData *[]Metrics) ([]byte, error) {
	var b bytes.Buffer

	w, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed init compress writer: %w", err)
	}

	metricDataBytes, err := json.Marshal(metricData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	_, err = w.Write(metricDataBytes)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %w", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %w", err)
	}

	return b.Bytes(), nil
}

// EncryptData - function for encrypting metricData
func EncryptData(data []byte, publicKeyStr []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKeyStr)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid PEM block type: %s", block.Type)
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub, data, nil)
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}

// DecryptData - function for decrypting data
func DecryptData(privateKeyStr string, data []byte) ([]byte, error) {

	block, _ := pem.Decode([]byte(privateKeyStr))
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, data, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func ConvertDataProtobuf(dataMetrics []Metrics) []*pb.Metric {
	protobufMetrics := make([]*pb.Metric, len(dataMetrics))
	for key, metric := range dataMetrics {
		var protobufMetric pb.Metric
		if metric.MType == "counter" {
			var counterValue pb.Metric_Delta
			counterValue.Delta = *metric.Delta
			protobufMetric = pb.Metric{Id: metric.ID, Mtype: pb.Metric_COUNTER, MetricValue: &counterValue}
		} else {
			var gaugeValue pb.Metric_Value
			gaugeValue.Value = *metric.Value
			protobufMetric = pb.Metric{Id: metric.ID, Mtype: pb.Metric_GAUGE, MetricValue: &gaugeValue}
		}

		protobufMetrics[key] = &protobufMetric
	}
	return protobufMetrics
}
