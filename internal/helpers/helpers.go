package helpers

import (
	"context"
	"encoding/json"
	"github.com/go-playground/form"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"net/http"
	"os"
)

type contextKey string

const UserIDKey contextKey = "userID"

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

func GetEnv(env string) (string, error) {
	if err := godotenv.Load(); err != nil {
		return "", err
	}
	port := os.Getenv(env)
	return port, nil
}

func ParseForm(r *http.Request, dest any) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	decoder := form.NewDecoder()
	if err := decoder.Decode(dest, r.PostForm); err != nil {
		return err
	}

	return nil
}

// InvalidCredentialsResponse todo: InvalidCredentialsResponse need to confirm that it's working
func InvalidCredentialsResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid credentials"})
}
