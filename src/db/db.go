package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"combina/src/types"
	"github.com/jackc/pgx/v4/pgxpool"
)

const defaultConnectionURL = "postgres://localhost:5432"

func connectionURL() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return defaultConnectionURL
}

func DatabaseConnect(dbName string) (*pgxpool.Pool, error) {
	conn, err := pgxpool.Connect(context.Background(), connectionURL()+dbName)
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
		if strings.Contains(err.Error(), "already exists") {
			log.Printf("Database already exists, skipping.")
			return nil
		}
		return err
	}

	log.Printf("Database created!")
	return nil
}

func createTable(conn *pgxpool.Pool) error {
	defer conn.Close()

	table := `CREATE TABLE IF NOT EXISTS lotto(
		id VARCHAR (50) PRIMARY KEY,
		type VARCHAR (50) NOT NULL,
		combination JSON NOT NULL,
		created_on TIMESTAMPTZ NOT NULL,
		name VARCHAR (50))`

	_, err := conn.Exec(context.Background(), table)
	if err != nil {
		return err
	}

	log.Printf("Table ready!")
	return nil
}
