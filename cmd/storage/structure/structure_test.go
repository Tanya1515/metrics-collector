package storage

import (
	"testing"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
	_ "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type InMemoryStorageSuite struct {
	suite.Suite
	Storage *MemStorage
}

func (MS *InMemoryStorageSuite) SetupSuite() {
	MS.Storage = &MemStorage{}
	err := MS.Storage.Init(false, "", 0)

	MS.NoError(err)
}

func TestInMemoryStorage(t *testing.T) {
	suite.Run(t, new(InMemoryStorageSuite))
}

func (MS *InMemoryStorageSuite) TestRepositoryAddValue() {
	MS.NoError(MS.Storage.RepositoryAddValue("TestCount", 100))

	testCounterValue, err := MS.Storage.GetCounterValueByName("TestCount")
	MS.NoError(err)

	MS.Equal(int64(100), testCounterValue)
}

func (MS *InMemoryStorageSuite) TestRepositoryAddCounterValue() {
	MS.NoError(MS.Storage.RepositoryAddCounterValue("TestCounter", 101))

	CounterValues, err := MS.Storage.GetAllCounterMetrics()
	MS.NoError(err)

	MS.Equal(int64(101), CounterValues["TestCounter"])
}

func (MS *InMemoryStorageSuite) TestRepositoryAddGaugeValue() {
	MS.NoError(MS.Storage.RepositoryAddGaugeValue("TestGauge", 101.101))

	GaugeValues, err := MS.Storage.GetAllGaugeMetrics()
	MS.NoError(err)

	MS.Equal(101.101, GaugeValues["TestGauge"])
}

func (MS *InMemoryStorageSuite) TestRepositoryAddAllValues() {
	metrics := make([]data.Metrics, 2)
	var testCounterAllDelta int64 = 101
	var testGaugeAllValue float64 = 101.101
	metrics[0] = data.Metrics{ID: "TestCounterAll", MType: "counter", Delta: &testCounterAllDelta}
	metrics[1] = data.Metrics{ID: "TestGaugeAll", MType: "gauge", Value: &testGaugeAllValue}

	MS.NoError(MS.Storage.RepositoryAddAllValues(metrics))

	counterRes, err :=MS.Storage.GetCounterValueByName("TestCounterAll")
	MS.NoError(err)
	MS.Equal(testCounterAllDelta, counterRes)

	gaugeRes, err := MS.Storage.GetGaugeValueByName("TestGaugeAll")
	MS.NoError(err)
	MS.Equal(testGaugeAllValue, gaugeRes)
}
