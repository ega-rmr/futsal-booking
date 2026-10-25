package db

import (
	"database/sql"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect() (*sql.DB, error) {
	conn, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	err = conn.Ping()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func CreateTables(conn *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		nama TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS courts (
		id SERIAL PRIMARY KEY,
		nama TEXT NOT NULL,
		harga_per_jam INT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS bookings (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL REFERENCES users(id),
		court_id INT NOT NULL REFERENCES courts(id),
		mulai TIMESTAMPTZ NOT NULL,
		selesai TIMESTAMPTZ NOT NULL,
		status TEXT NOT NULL DEFAULT 'aktif'
	);`

	_, err := conn.Exec(query)
	return err
}