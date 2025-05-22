package helpers

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"os"
)

func OpenDB() (*pgx.Conn, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}
	dbUrl := os.Getenv("DB_URL")
	conn, err := pgx.Connect(context.Background(), dbUrl)
	if err != nil {
		return nil, err
	}
	err = conn.Ping(context.Background())
	if err != nil {
		_ = conn.Close(context.Background())
		return nil, err
	}

	return conn, nil
}

func GetPort(portName string) (string, error) {
	if err := godotenv.Load(); err != nil {
		return "", err
	}
	port := os.Getenv(portName)
	return port, nil
}
