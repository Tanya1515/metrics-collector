package storage

import (
	"database/sql"
	
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgreSQLConnection struct{
	Address string
	UserName string
	Password string
	dbConn   *sql.DB
} 

func (db *PostgreSQLConnection) InitConnection() (*sql.DB, error){
	return nil, nil
}

func (db *PostgreSQLConnection) CheckConnection() error {

	return nil
}