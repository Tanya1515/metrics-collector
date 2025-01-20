package storage

import (
	"errors"
	"fmt"

	sql "database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func (db *PostgreSQLConnection) RepositoryAddCounterValue(metricName string, metricValue int64) error {
	var value int64
	row := db.dbConn.QueryRow(`SELECT Value FROM `+MetricsTableName+` WHERE metricType=? AND metricName=?;`, "counter", metricName)

	err := row.Scan(&value)
	if (err != nil) && !(errors.Is(err, sql.ErrNoRows)) {
		return fmt.Errorf("error while getting gauge metric value %w with name %s", err, metricName)
	}

	tx, err := db.dbConn.Begin()
	if err != nil {
		return fmt.Errorf("error while starting transaction: %w", err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		_, err = tx.Exec(
			"INSERT INTO"+MetricsTableName+" (metricType, metricName, Delta)"+
				" VALUES(?,?,?)", "counter", metricName, metricValue)
	} else {
		value = value + metricValue
		_, err = tx.Exec(
			"UPDATE "+MetricsTableName+" SET Delta=? WHERE metricName=? AND metricType=?;", value, metricName, "counter")
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
	var value int64
	row := db.dbConn.QueryRow(`SELECT Delta FROM `+MetricsTableName+` WHERE metricType=? AND metricName=?;`, "gauge", metricName)

	err := row.Scan(&value)
	if (err != nil) && !(errors.Is(err, sql.ErrNoRows)) {
		return fmt.Errorf("error while getting gauge metric value %w with name %s", err, metricName)
	}

	tx, err := db.dbConn.Begin()
	if err != nil {
		return fmt.Errorf("error while starting transaction: %w", err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		_, err = tx.Exec(
			"INSERT INTO "+MetricsTableName+" (metricType, metricName, Value)"+
				" VALUES(?,?,?)", "gauge", metricName, metricValue)
	} else {
		_, err = tx.Exec(
			"UPDATE "+MetricsTableName+" SET Value=? WHERE metricName=? AND metricType=?;", metricValue, metricName, "gauge")
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
	row := db.dbConn.QueryRow(`SELECT Value FROM `+MetricsTableName+` WHERE metricType=? AND metricName=?;`, "counter", metricName)

	err := row.Scan(&value)
	if (err != nil) && !(errors.Is(err, sql.ErrNoRows)) {
		return fmt.Errorf("error while getting gauge metric value %w with name %s", err, metricName)
	}

	tx, err := db.dbConn.Begin()
	if err != nil {
		return fmt.Errorf("error while starting transaction: %w", err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		_, err = tx.Exec(
			"INSERT INTO"+MetricsTableName+" (metricType, metricName, Delta)"+
				" VALUES(?,?,?);", "counter", metricName, metricValue)
	} else {
		_, err = tx.Exec(
			"UPDATE "+MetricsTableName+" SET Delta=? WHERE metricName=? AND metricType=?;", value, metricName, "counter")
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
