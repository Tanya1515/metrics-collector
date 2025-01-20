package storage

import (
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// если метрика существует - проверять по имени и выполнять update
// если метрики нет - выполнять insert into

func (db *PostgreSQLConnection) RepositoryAddCounterValue(metricName string, metricValue int64) error {

	tx, err := db.dbConn.Begin()
	if err != nil {
		return fmt.Errorf("error while starting transaction: %w", err)
	}

	_, err = tx.Exec(
		"INSERT INTO"+MetricsTableName+"(metricType, metricName, Delta)"+
			" VALUES(?,?,?)", "counter", metricName, metricValue)

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

	tx, err := db.dbConn.Begin()
	if err != nil {
		return fmt.Errorf("error while starting transaction: %w", err)
	}

	_, err = tx.Exec(
		"INSERT INTO"+MetricsTableName+"(metricType, metricName, Value)"+
			" VALUES(?,?,?)", "gauge", metricName, metricValue)

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
	return nil
}
