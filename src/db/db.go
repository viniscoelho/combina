package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"combina/src/types"
	"github.com/jackc/pgx/v4/pgxpool"
)

/*
	DATABASE_URL:               postgres://{user}:{password}@{hostname}:{port}/{database-name}
*/

const connectionURL = "postgres://localhost:5432"

func DatabaseConnect(dbName string) (*pgxpool.Pool, error) {
	err := os.Setenv("DATABASE_URL", fmt.Sprintf("%s%s", connectionURL, dbName))
	if err != nil {
		return nil, fmt.Errorf("failed to set DATABASE_URL: %w", err)
	}

	conn, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return conn, err
}

// call this if a flag is passed in the main function
func InitializeDatabase() {
	conn, err := DatabaseConnect("")
	if err != nil {
		log.Fatalf("An error occured: %s", err)
	}

	if err = createDatabase(conn); err != nil {
		log.Fatalf("Database creation failed: %s", err)
	}

	conn, err = DatabaseConnect(types.DatabaseName)
	if err != nil {
		log.Fatalf("An error occured: %s", err)
	}

	if err = createTable(conn); err != nil {
		log.Fatalf("Table creation failed: %s", err)
	}
}

func createDatabase(conn *pgxpool.Pool) error {
	defer conn.Close()

	_, err := conn.Exec(context.Background(), `CREATE DATABASE lotto`)
	if err != nil {
		return err
	}

	log.Printf("Database created!")
	return nil
}

func createTable(conn *pgxpool.Pool) error {
	defer conn.Close()

	table := `CREATE TABLE lotto(
		id VARCHAR (50) PRIMARY KEY,
		type VARCHAR (50) NOT NULL,
		combination JSON NOT NULL,
		created_on TIMESTAMPTZ NOT NULL,
		name VARCHAR (50))`

	_, err := conn.Exec(context.Background(), table)
	if err != nil {
		return err
	}

	log.Printf("Table created!")
	return nil
}
