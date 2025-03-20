package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeString(t *testing.T) {
	tests := []struct {
		name          string
		serverAddress string
		result        string
	}{
		{
			name:          "test: Make http-request with gauge metric",
			serverAddress: "localhost:8080",
			result:        "http://localhost:8080/updates/",
		},
		{
			name:          "test: Make http-request with counter metric",
			serverAddress: "192.168.0.1:8085",
			result:        "http://192.168.0.1:8085/updates/",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := MakeString(test.serverAddress)
			assert.Equal(t, test.result, result)
		})
	}
}
