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

// Парсим публичный ключ из DER-данных

// Проверяем, что это RSA ключ

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
