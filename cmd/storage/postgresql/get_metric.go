package storage

import (
	_ "database/sql"
	
	_ "github.com/jackc/pgx/v5/stdlib"
)

func (db *PostgreSQLConnection) GetCounterValueByName(metricName string) (int64, error) {
	return 0, nil
}

func (db *PostgreSQLConnection) GetGaugeValueByName(metricName string) (float64, error) {
	return 0, nil
}
