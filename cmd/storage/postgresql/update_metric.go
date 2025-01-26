package storage

import (
	"errors"
	"fmt"

	sql "database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func (db *PostgreSQLConnection) RepositoryAddCounterValue(metricName string, metricValue int64) error {
	var value int64
	row := db.dbConn.QueryRow("SELECT Delta FROM "+MetricsTableName+" WHERE metricType = $1 AND metricName = $2", "counter", metricName)

	err := row.Scan(&value)
	if (err != nil) && !(errors.Is(err, sql.ErrNoRows)) {
		return fmt.Errorf("error while getting gauge metric value %w with name %s", err, metricName)
	}

	tx, errTr := db.dbConn.Begin()
	if errTr != nil {
		return fmt.Errorf("error while starting transaction: %w", err)
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
	row := db.dbConn.QueryRow("SELECT Value FROM "+MetricsTableName+" WHERE metricType = $1 AND metricName = $2", "gauge", metricName)

	err := row.Scan(&value)
	if (err != nil) && !(errors.Is(err, sql.ErrNoRows)) {
		return fmt.Errorf("error while getting gauge metric value %w", err)
	}

	tx, errTr := db.dbConn.Begin()
	if errTr != nil {
		return fmt.Errorf("error while starting transaction: %w", err)
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
	row := db.dbConn.QueryRow("SELECT Delta FROM "+MetricsTableName+" WHERE metricType = $1 AND metricName = $2", "counter", metricName)

	err := row.Scan(&value)
	if (err != nil) && !(errors.Is(err, sql.ErrNoRows)) {
		return fmt.Errorf("error while getting counter metric value %w", err)
	}

	tx, errTr := db.dbConn.Begin()
	if errTr != nil {
		return fmt.Errorf("error while starting transaction: %w", err)
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
