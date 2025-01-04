package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckValue(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		result    bool
	}{
		{
			name:      "test: Name of Metric exists",
			fieldName: "Alloc",
			result:    true,
		},
		{
			name:      "test: Name of Metric does not exist",
			fieldName: "Test",
			result:    false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := CheckValue(test.fieldName)
			assert.Equal(t, test.result, result)
		})
	}
}

func TestMakeString(t *testing.T) {
	tests := []struct {
		name          string
		serverAddress string
		result        string
	}{
		{
			name:          "test: Make http-request with gauge metric",
			serverAddress: "localhost:8080",
			result:        "http://localhost:8080/update",
		},
		{
			name:          "test: Make http-request with counter metric",
			serverAddress: "192.168.0.1:8085",
			result:        "http://192.168.0.1:8085/update",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := MakeString(test.serverAddress)
			assert.Equal(t, test.result, result)
		})
	}
}
