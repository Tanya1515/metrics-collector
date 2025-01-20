package storage

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgreSQLConnection struct {
	Address  string
	UserName string
	Password string
	DBName   string
	dbConn   *sql.DB
}

<<<<<<< HEAD
func (db *PostgreSQLConnection) Init(restore bool, fileStore string, backupTimer int) error {
=======
const (
	MetricsTableName = "metrics"
)

// подумать, что делать с контекстом

func (db *PostgreSQLConnection) Init() error {
>>>>>>> d454a90 (add creation of table in postgresql and select requests to postgresql)
	var err error
	ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		db.Address, db.UserName, db.Password, db.DBName)

	db.dbConn, err = sql.Open("pgx", ps)
	if err != nil {
		return err
	}

	_, err = db.dbConn.Exec(`CREATE TABLE ` + MetricsTableName + ` (Id INTEGER PRIMARY KEY,
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
