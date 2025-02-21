package storage

import (
	sql "database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
)

func (db *PostgreSQLConnection) RepositoryAddCounterValue(metricName string, metricValue int64) error {
	var value int64

	tx, err := db.dbConn.Begin()

	if err != nil {
		return fmt.Errorf("error while starting transaction: %w", err)
	}

	row := tx.QueryRow("SELECT Delta FROM "+MetricsTableName+" WHERE metricType = $1 AND metricName = $2 FOR UPDATE", "counter", metricName)

	err = row.Scan(&value)
	if (err != nil) && !(errors.Is(err, sql.ErrNoRows)) {
		tx.Rollback()
		return fmt.Errorf("error while getting gauge metric value %w with name %s", err, metricName)
	}

	_, err = tx.Exec("INSERT INTO "+MetricsTableName+" (metricType, metricName, Delta) VALUES ($1,$2,$3)"+
		" ON CONFLICT (metricName) DO"+
		" UPDATE SET Delta = excluded.Delta WHERE metrics.metricType = excluded.metricType AND metrics.metricName = excluded.metricName", "counter", metricName, metricValue+value)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error while adding counter metric with name %s:  %w", metricName, err)
	}
	err = tx.Commit()

	if err != nil {
		return fmt.Errorf("error while closing transaction: %w", err)
	}
	return nil
}

func (db *PostgreSQLConnection) RepositoryAddGaugeValue(metricName string, metricValue float64) error {

	_, err := db.dbConn.Exec("INSERT INTO "+MetricsTableName+" (metricType, metricName, Value) VALUES($1,$2,$3)"+
		" ON CONFLICT (metricName) DO"+
		" UPDATE SET Value = EXCLUDED.Value WHERE metrics.metricType = EXCLUDED.metricType AND metrics.metricName = EXCLUDED.metricName", "gauge", metricName, metricValue)

	if err != nil {
		return fmt.Errorf("error during adding new gauge metricValue: %w", err)
	}
	return nil
}

func (db *PostgreSQLConnection) RepositoryAddValue(metricName string, metricValue int64) error {

	_, err := db.dbConn.Exec("INSERT INTO "+MetricsTableName+" (metricType, metricName, Delta) VALUES ($1,$2,$3) "+
		" ON CONFLICT (metricName) DO"+
		" UPDATE SET Delta = excluded.Delta WHERE metrics.metricType = excluded.metricType AND metrcis.metricName = excluded.metricName", "counter", metricName, metricValue)

	if err != nil {
		return fmt.Errorf("error during adding new counter metricValue: %w", err)
	}
	return nil
}

func (db *PostgreSQLConnection) RepositoryAddAllValues(metrics []data.Metrics) error {

	var valueCounter int64
	tx, err := db.dbConn.Begin()

	if err != nil {
		return fmt.Errorf("error while starting transaction: %w", err)
	}

	for _, metric := range metrics {
		if metric.MType == "counter" {
			row := tx.QueryRow("SELECT Delta FROM "+MetricsTableName+" WHERE metricType = $1 AND metricName = $2 FOR UPDATE", metric.MType, metric.ID)

			err := row.Scan(&valueCounter)
			if (err != nil) && !(errors.Is(err, sql.ErrNoRows)) {
				tx.Rollback()
				return fmt.Errorf("error while getting counter metric value %w", err)
			}
			_, err = tx.Exec("INSERT INTO "+MetricsTableName+" (metricType, metricName, Delta) VALUES ($1,$2,$3)"+
				" ON CONFLICT (metricName) DO"+
				" UPDATE SET Delta = excluded.Delta WHERE metrics.metricType = excluded.metricType AND metrics.metricName = excluded.metricName", metric.MType, metric.ID, *metric.Delta+valueCounter)

			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error while updating counter metric with name %s:  %w", metric.ID, err)
			}
		} else if metric.MType == "gauge" {

			_, err := tx.Exec("INSERT INTO "+MetricsTableName+" (metricType, metricName, Value) VALUES ($1,$2,$3)"+
				" ON CONFLICT (metricName) DO"+
				" UPDATE SET Value = excluded.Value WHERE metrics.metricType = excluded.metricType AND metrics.metricName = excluded.metricName", metric.MType, metric.ID, *metric.Value)

			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error while adding gauge metric with name %s:  %w", metric.ID, err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error while closing transaction: %w", err)
	}
	return nil
}
