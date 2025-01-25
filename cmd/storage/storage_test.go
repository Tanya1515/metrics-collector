package storage

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Tanya1515/metrics-collector.git/cmd/mocks"
)

func TestGetCounterValueByName(t *testing.T) {

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRepositoryInterface(ctrl)
	var value int64 = 5
	m.EXPECT().GetCounterValueByName("TestCount").Return(value, nil)

	val, err := m.GetCounterValueByName("TestCount")

	require.NoError(t, err)
	require.Equal(t, val, value)

}

func TestGetGaugeValueByName(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRepositoryInterface(ctrl)
	value := 7.3
	m.EXPECT().GetGaugeValueByName("TestGauge").Return(value, nil)

	val, err := m.GetGaugeValueByName("TestGauge")

	require.NoError(t, err)
	require.Equal(t, val, value)
}

func TestGetAllGaugeMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRepositoryInterface(ctrl)

	AllGaugeMetrics := make(map[string]float64, 3)
	AllGaugeMetrics["TestGauge_1"] = 1.1
	AllGaugeMetrics["TestGauge_2"] = 1.2
	AllGaugeMetrics["TestGauge_3"] = 1.3

	m.EXPECT().GetAllGaugeMetrics().Return(AllGaugeMetrics, nil)

	res, err := m.GetAllGaugeMetrics()

	require.NoError(t, err)
	require.Equal(t, res, AllGaugeMetrics)
}

func TestGetAllCounterMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRepositoryInterface(ctrl)

	AllCounterMetrics := make(map[string]float64, 3)
	AllCounterMetrics["TestCounter_1"] = 1
	AllCounterMetrics["TestCounter_2"] = 2
	AllCounterMetrics["TestCounter_3"] = 3

	m.EXPECT().GetAllGaugeMetrics().Return(AllCounterMetrics, nil)

	res, err := m.GetAllGaugeMetrics()

	require.NoError(t, err)
	require.Equal(t, res, AllCounterMetrics)
}

func TestRepositoryAddCounterValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRepositoryInterface(ctrl)
	var value int64 = 5
	m.EXPECT().RepositoryAddCounterValue("TestCounter", value).Return(nil)

	err := m.RepositoryAddCounterValue("TestCounter", value)

	require.NoError(t, err)
}

func TestRepositoryAddGaugeValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRepositoryInterface(ctrl)
	value := 1.0
	m.EXPECT().RepositoryAddGaugeValue("TestGauge", value).Return(nil)

	err := m.RepositoryAddGaugeValue("TestGauge", value)

	require.NoError(t, err)
}

func TestRepositoryAddValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRepositoryInterface(ctrl)
	var value int64 = 5
	m.EXPECT().RepositoryAddValue("Test", value).Return(nil)

	err := m.RepositoryAddValue("Test", value)

	require.NoError(t, err)
}
