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

func (db *PostgreSQLConnection) Init() error {
	var err error
	ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		db.Address, db.UserName, db.Password, db.DBName)

	db.dbConn, err = sql.Open("pgx", ps)
	if err != nil {
		return err
	}
	return nil
}

func (db *PostgreSQLConnection) CheckConnection(ctx context.Context) error {

	return db.dbConn.PingContext(ctx)
}
