package storage

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgreSQLConnection struct {
	Address  string
	Port     string
	UserName string
	Password string
	DBName   string
	dbConn   *sql.DB
}

const (
	MetricsTableName = "metrics"
)

func (db *PostgreSQLConnection) Init(restore bool, fileStore string, backupTimer int) error {
	var err error
	ps := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		db.Address, db.Port, db.UserName, db.Password, db.DBName)

	db.dbConn, err = sql.Open("pgx", ps)
	if err != nil {
		return err
	}

	_, err = db.dbConn.Exec(`CREATE TABLE ` + MetricsTableName + ` (Id BIGSERIAL PRIMARY KEY,
	                                                                metricName VARCHAR(100) NOT NULL,
																	metricType VARCHAR(100) NOT NULL,
																	Delta INTEGER, 
																	Value DOUBLE PRECISION);`)

	if err != nil {
		return fmt.Errorf("error %w occured while creating table %s", err, MetricsTableName)
	}

	return nil
}

func (db *PostgreSQLConnection) CheckConnection(ctx context.Context) error {

	return db.dbConn.PingContext(ctx)
}
