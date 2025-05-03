package storage

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
)

type PostgresTestSuite struct {
	suite.Suite
	QueryTimeout time.Duration
	tc           *tcpostgres.PostgresContainer
	cfg          *PostgreSQLConnection
}

func (ts *PostgresTestSuite) SetupSuite() {

	cfg := &PostgreSQLConnection{
		UserName: "postgres",
		Password: "postgres",
		DBName:   "postgres",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// create container with docker image for postgresql with database
	pgc, err := tcpostgres.Run(ctx,
		"docker.io/postgres:latest",
		tcpostgres.WithDatabase(cfg.DBName),
		tcpostgres.WithUsername(cfg.UserName),
		tcpostgres.WithPassword(cfg.Password),
		// wait for the condition
		// in the case wait for the string from log
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(5*time.Second)),
	)

	// check if no error arrives
	require.NoError(ts.T(), err)

	cfg.Address, err = pgc.Host(ctx)
	require.NoError(ts.T(), err)

	//get mapped port from container to external
	port, err := pgc.MappedPort(ctx, "5432")
	require.NoError(ts.T(), err)

	cfg.Port = strconv.Itoa((port.Int()))

	ts.tc = pgc
	ts.cfg = cfg
	ts.QueryTimeout = 5 * time.Second
	chanSh := make(chan struct{})
	require.NoError(ts.T(), cfg.Init(false, "", 0, chanSh, context.Background()))

	ts.T().Logf("started postgres at %s:%s", ts.cfg.Address, ts.cfg.Port)
}

// remove container with postgres
func (ts *PostgresTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	require.NoError(ts.T(), ts.tc.Terminate(ctx))
}

// clean all tables in database for running separate tests
func (ts *PostgresTestSuite) clean(ctx context.Context) error {
	newctx, cancel := context.WithTimeout(ctx, ts.QueryTimeout)
	defer cancel()

	_, err := ts.cfg.dbConn.ExecContext(newctx, "DELETE FROM metrics")
	return err
}

// function, that is running before every test case
func (ts *PostgresTestSuite) SetupTest() {
	ts.Require().NoError(ts.clean(context.Background()))
}

// function, that is applied after test case for cleaning up test environment
func (ts *PostgresTestSuite) TearDownTest() {
	ts.Require().NoError(ts.clean(context.Background()))
}

func TestPostgres(t *testing.T) {
	suite.Run(t, new(PostgresTestSuite))
}

func (ts *PostgresTestSuite) TestRepositoryAddCounterValue() {
	ts.NoError(ts.cfg.RepositoryAddCounterValue("TestCounter", 100))

	counterMetrics, err := ts.cfg.GetAllCounterMetrics()
	ts.NoError(err)

	ts.Contains(counterMetrics, "TestCounter")
}

func (ts *PostgresTestSuite) TestRepositoryAddAllValues() {
	metrics := make([]data.Metrics, 2)
	var testCounterAllDelta int64 = 101
	testGaugeAllValue := 101.101
	metrics[0] = data.Metrics{ID: "TestCounterAll", MType: "counter", Delta: &testCounterAllDelta}
	metrics[1] = data.Metrics{ID: "TestGaugeAll", MType: "gauge", Value: &testGaugeAllValue}

	ts.NoError(ts.cfg.RepositoryAddAllValues(metrics))

	counterRes, err := ts.cfg.GetCounterValueByName("TestCounterAll")
	ts.NoError(err)
	ts.Equal(testCounterAllDelta, counterRes)

	gaugeRes, err := ts.cfg.GetGaugeValueByName("TestGaugeAll")
	ts.NoError(err)
	ts.Equal(testGaugeAllValue, gaugeRes)

}

func (ts *PostgresTestSuite) TestRepositoryAddGaugeValue() {
	ts.NoError(ts.cfg.RepositoryAddGaugeValue("TestGauge", 101.101))

	gaugeMetrics, err := ts.cfg.GetAllGaugeMetrics()
	ts.NoError(err)

	ts.Contains(gaugeMetrics, "TestGauge")
	ts.Equal(101.101, gaugeMetrics["TestGauge"])
}

func (ts *PostgresTestSuite) TestRepositoryAddValue() {
	ts.NoError(ts.cfg.RepositoryAddValue("TestCounter", 100))

	counterMetrics, err := ts.cfg.GetAllCounterMetrics()
	ts.NoError(err)

	ts.Contains(counterMetrics, "TestCounter")
}

func (ts *PostgresTestSuite) TestCheckConnection() {
	newctx, cancel := context.WithTimeout(context.Background(), ts.QueryTimeout)
	defer cancel()
	ts.NoError(ts.cfg.CheckConnection(newctx))
}
