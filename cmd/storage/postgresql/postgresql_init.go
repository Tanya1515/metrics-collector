package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	storage "github.com/Tanya1515/metrics-collector.git/cmd/storage"
)

type PostgreSQLConnection struct {
	storage.StoreType
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

func (db *PostgreSQLConnection) Init(ctx context.Context, shutdown chan struct{}) error {
	var err error

	ps := fmt.Sprintf("host=%s port=%s user=%s password=%s database=%s sslmode=disable",
		db.Address, db.Port, db.UserName, db.Password, db.DBName)

	db.dbConn, err = sql.Open("pgx", ps)
	if err != nil {
		return err
	}

	_, err = db.dbConn.Exec(`CREATE TABLE ` + MetricsTableName + ` (Id BIGSERIAL PRIMARY KEY,
	                                                                metricName VARCHAR(100) NOT NULL UNIQUE,
																	metricType VARCHAR(100) NOT NULL,
																	Delta BIGINT, 
																	Value DOUBLE PRECISION);`)
	if err != nil {
		return err
	}

	db.dbConn.Close()

	return nil
}

func (db *PostgreSQLConnection) CheckConnection(ctx context.Context) error {

	return db.dbConn.PingContext(ctx)
}

func (db *PostgreSQLConnection) CloseConnections() error {
	err := db.dbConn.Close()
	if err != nil {
		return err
	}

	close(db.Shutdown)
	return nil
}
