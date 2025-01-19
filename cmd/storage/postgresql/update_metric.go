package storage

import (
	_ "database/sql"
	
	_ "github.com/jackc/pgx/v5/stdlib"
)


func RepositoryAddCounterValue(metricName string, metricValue int64) {

}
func RepositoryAddGaugeValue(metricName string, metricValue float64) {

}

func RepositoryAddValue(metricName string, metricValue int64) {

}