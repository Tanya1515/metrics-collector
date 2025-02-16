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

	row := tx.QueryRow("SELECT Delta FROM "+MetricsTableName+" WHERE metricType = $1 AND metricName = $2", "counter", metricName)

	err = row.Scan(&value)
	if (err != nil) && !(errors.Is(err, sql.ErrNoRows)) {
		return fmt.Errorf("error while getting gauge metric value %w with name %s", err, metricName)
	}

	if errors.Is(err, sql.ErrNoRows) {
		_, err = tx.Exec(
			"INSERT INTO "+MetricsTableName+" (metricType, metricName, Delta)"+
				" VALUES($1,$2,$3)", "counter", metricName, metricValue)
	} else {
		value = value + metricValue
		_, err = tx.Exec(
			"UPDATE "+MetricsTableName+" SET Delta = $1 WHERE metricName = $2 AND metricType = $3", value, metricName, "counter")
	}

	if err != nil {
		// если ошибка, то откатываем изменения
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
	var value float64

	tx, err := db.dbConn.Begin()

	if err != nil {
		return fmt.Errorf("error while starting transaction: %w", err)
	}

	row := tx.QueryRow("SELECT Value FROM "+MetricsTableName+" WHERE metricType = $1 AND metricName = $2", "gauge", metricName)

	err = row.Scan(&value)
	if (err != nil) && !(errors.Is(err, sql.ErrNoRows)) {
		return fmt.Errorf("error while getting gauge metric value %w", err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		_, err = tx.Exec(
			"INSERT INTO "+MetricsTableName+" (metricType, metricName, Value)"+
				" VALUES($1,$2,$3)", "gauge", metricName, metricValue)
	} else {
		_, err = tx.Exec(
			"UPDATE "+MetricsTableName+" SET Value = $1 WHERE metricName = $2 AND metricType = $3", metricValue, metricName, "gauge")
	}

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error while adding gauge metric with name %s:  %w", metricName, err)
	}
	err = tx.Commit()

	if err != nil {
		return fmt.Errorf("error while closing transaction: %w", err)
	}
	return nil
}

func (db *PostgreSQLConnection) RepositoryAddValue(metricName string, metricValue int64) error {
	var value int64

	tx, errTr := db.dbConn.Begin()

	if errTr != nil {
		return fmt.Errorf("error while starting transaction: %w", errTr)
	}

	row := tx.QueryRow("SELECT Delta FROM "+MetricsTableName+" WHERE metricType = $1 AND metricName = $2", "counter", metricName)

	err := row.Scan(&value)
	if (err != nil) && !(errors.Is(err, sql.ErrNoRows)) {
		return fmt.Errorf("error while getting counter metric value %w", err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		_, err = tx.Exec(
			"INSERT INTO "+MetricsTableName+" (metricType, metricName, Delta)"+
				" VALUES($1,$2,$3);", "counter", metricName, metricValue)
	} else {
		_, err = tx.Exec(
			"UPDATE "+MetricsTableName+" SET Delta = $1 WHERE metricName = $2 AND metricType = $3", value, metricName, "counter")
	}

	if err != nil {
		// если ошибка, то откатываем изменения
		tx.Rollback()
		return fmt.Errorf("error while adding counter metric with name %s:  %w", metricName, err)
	}
	err = tx.Commit()

	if err != nil {
		return fmt.Errorf("error while closing transaction: %w", err)
	}
	return nil
}

func (db *PostgreSQLConnection) RepositoryAddAllValues(metrics []data.Metrics) error {

	var valueCounter int64
	var valueGauge float64
	tx, err := db.dbConn.Begin()

	if err != nil {
		return fmt.Errorf("error while starting transaction: %w", err)
	}

	for _, metric := range metrics {
		if metric.MType == "counter" {
			row := tx.QueryRow("SELECT Delta FROM "+MetricsTableName+" WHERE metricType = $1 AND metricName = $2", metric.MType, metric.ID)

			err := row.Scan(&valueCounter)
			if (err != nil) && !(errors.Is(err, sql.ErrNoRows)) {
				tx.Rollback()
				return fmt.Errorf("error while getting counter metric value %w", err)
			}

			if errors.Is(err, sql.ErrNoRows) {
				_, err = tx.Exec(
					"INSERT INTO "+MetricsTableName+" (metricType, metricName, Delta)"+
						" VALUES($1,$2,$3);", "counter", metric.ID, *metric.Delta)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("error while inserting counter metric with name %s:  %w", metric.ID, err)
				}
			} else {
				_, err = tx.Exec(
					"UPDATE "+MetricsTableName+" SET Delta = $1 WHERE metricName = $2 AND metricType = $3", valueCounter+*metric.Delta, metric.ID, "counter")

				if err != nil {
					tx.Rollback()
					return fmt.Errorf("error while updating counter metric with name %s:  %w", metric.ID, err)
				}
			}

		} else if metric.MType == "gauge" {
			row := tx.QueryRow("SELECT Value FROM "+MetricsTableName+" WHERE metricType = $1 AND metricName = $2", metric.MType, metric.ID)

			err := row.Scan(&valueGauge)
			if (err != nil) && !(errors.Is(err, sql.ErrNoRows)) {
				tx.Rollback()
				return fmt.Errorf("error while getting gauge metric value %w", err)
			}

			if errors.Is(err, sql.ErrNoRows) {
				_, err = tx.Exec(
					"INSERT INTO "+MetricsTableName+" (metricType, metricName, Value)"+
						" VALUES($1,$2,$3)", "gauge", metric.ID, *metric.Value)
			} else {
				_, err = tx.Exec(
					"UPDATE "+MetricsTableName+" SET Value = $1 WHERE metricName = $2 AND metricType = $3", *metric.Value, metric.ID, "gauge")
			}
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
