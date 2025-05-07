package postgresql

import (
	"context"
	"database/sql"
	"fmt"

	storage "github.com/Tanya1515/metrics-collector.git/cmd/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
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

func (db *PostgreSQLConnection) Init(shutdown chan struct{}, ctx context.Context) error {
	var err error
	if db.Restore {
		err := db.Store(db)
		if err != nil {
			return err
		}
	}

	if (db.FileStore != "") && (db.BackupTimer != 0) {

		go db.SaveMetricsAsync(shutdown, ctx, db)
	}

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

	return nil
}

func (db *PostgreSQLConnection) CheckConnection(ctx context.Context) error {

	return db.dbConn.PingContext(ctx)
}
