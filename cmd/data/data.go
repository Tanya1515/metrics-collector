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
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

// ResultMetrics - type, that contains all gauge and counter metrics for representation them in HTML-format.
type ResultMetrics struct {
	GaugeMetrics   string // All gauge metrics
	CounterMetrics string // All counter
}

// Metrics - type, that describes all fields of recieved/saved metrics.
type Metrics struct {
	ID    string   `json:"id"`              // Metric name
	MType string   `json:"type"`            // Metric Type (counter or gauge)
	Delta *int64   `json:"delta,omitempty"` // Counter value
	Value *float64 `json:"value,omitempty"` // Gauge Value
}

// ConfigApp - type, that describes all fields of the application config file
type ConfigApp struct {
	ServerAddress string `json:"address"`        // Server address
	StoreInterval string `json:"store_interval"` // Time duration for saving metrics
	FileStorePath string `json:"store_file"`     // Filename for storing metrics
	Restore       bool   `json:"restore"`        // Flag for storing all info
	PostgreSQL    string `json:"database_dsn"`   // Credentials for database
	SecretKey     string `json:"secret_key"`     // Secret key for hashing data
	CryptoKeyPath string `json:"crypto_key"`     // Path to key for asymmetrical encryption
	TrustedSubnet string `json:"trusted_subnet"` // Trusted IP address range
}

// ConfigAgent - type, that describes all fields of the agent configuration
type ConfigAgent struct {
	ReportInterval      string `json:"report_interval"` // Time duration for saving metrics
	PollInterval        string `json:"poll_interval"`   // Time duration for getting
	ServerAddress       string `json:"address"`         // Address for sending metrics
	SecretKey           string `json:"secret_key"`      // Secret hash for creating hash
	CryptoKeyPath       string `json:"crypto_key"`      // Requests linit for server
	LimitServerRequests int    `json:"limit_requests"`  // Key path for assymetrical encryption
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
