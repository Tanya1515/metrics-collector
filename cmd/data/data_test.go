package data

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompress(t *testing.T) {
	var buf bytes.Buffer
	metricsTest := make([]Metrics, 3)
	metricsResult := make([]Metrics, 2)
	var testCounterAllDelta int64 = 101
	testGaugeAllValue  := 101.101
	metricsTest[0] = Metrics{ID: "TestCounterAll", MType: "counter", Delta: &testCounterAllDelta}
	metricsTest[1] = Metrics{ID: "TestGaugeAll", MType: "gauge", Value: &testGaugeAllValue}

	t.Run("Test:", func(t *testing.T) {
		result, err := Compress(&metricsTest)
		require.NoError(t, err)

		gz, err := gzip.NewReader(bytes.NewReader(result))
		require.NoError(t, err)

		_, err = buf.ReadFrom(gz)
		require.NoError(t, err)

		err = json.Unmarshal(buf.Bytes(), &metricsResult)
		require.NoError(t, err)

		assert.Equal(t, "TestCounterAll", (metricsResult[0].ID))
		assert.Equal(t, "TestGaugeAll", (metricsResult[1].ID))
	})

}
