package postgresql

import (
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func (db *PostgreSQLConnection) GetCounterValueByName(metricName string) (delta int64, err error) {

	row := db.dbConn.QueryRow("SELECT Delta FROM "+MetricsTableName+" WHERE metricType = $1 AND metricName = $2", "counter", metricName)

	err = row.Scan(&delta)
	if err != nil {
		return 0, fmt.Errorf("error while getting counter metric value %w with name %s", err, metricName)
	}

	return
}

func (db *PostgreSQLConnection) GetGaugeValueByName(metricName string) (value float64, err error) {

	row := db.dbConn.QueryRow("SELECT Value FROM "+MetricsTableName+" WHERE metricType = $1 AND metricName = $2", "gauge", metricName)

	err = row.Scan(&value)
	if err != nil {
		return 0, fmt.Errorf("error while getting gauge metric value %w with name %s", err, metricName)
	}

	return
}

func (db *PostgreSQLConnection) GetAllGaugeMetrics() (map[string]float64, error) {

	gaugeMetrics := make(map[string]float64, 100)

	rows, err := db.dbConn.Query("SELECT metricName, Value FROM "+MetricsTableName+" WHERE metricType = $1", "gauge")
	if err != nil {
		return gaugeMetrics, fmt.Errorf("error while getting all gauge metrics: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var metricName string
		var metricValue float64
		err = rows.Scan(&metricName, &metricValue)
		if err != nil {
			return gaugeMetrics, fmt.Errorf("error while processing data: %w", err)
		}
		gaugeMetrics[metricName] = metricValue
	}
	err = rows.Err()
	if err != nil {
		return gaugeMetrics, fmt.Errorf("error while getting new data: %w", err)
	}

	return gaugeMetrics, nil
}

func (db *PostgreSQLConnection) GetAllCounterMetrics() (map[string]int64, error) {

	conterMetrics := make(map[string]int64, 100)

	rows, err := db.dbConn.Query("SELECT metricName, Delta FROM metrics WHERE metricType = $1", "counter")
	if err != nil {
		return conterMetrics, fmt.Errorf("error while getting all counter metrics: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var metricName string
		var metricDelta int64

		err = rows.Scan(&metricName, &metricDelta)
		if err != nil {
			return conterMetrics, fmt.Errorf("error while processing data: %w", err)
		}
		conterMetrics[metricName] = metricDelta
	}

	err = rows.Err()
	if err != nil {
		return conterMetrics, fmt.Errorf("error while getting new data: %w", err)
	}

	return conterMetrics, nil
}
