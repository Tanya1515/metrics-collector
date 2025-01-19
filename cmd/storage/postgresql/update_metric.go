package storage

import (
	_ "database/sql"
	
	_ "github.com/jackc/pgx/v5/stdlib"
)


func (db *PostgreSQLConnection) RepositoryAddCounterValue(metricName string, metricValue int64) {

}
func (db *PostgreSQLConnection) RepositoryAddGaugeValue(metricName string, metricValue float64) {

}

func (db *PostgreSQLConnection) RepositoryAddValue(metricName string, metricValue int64) {

}