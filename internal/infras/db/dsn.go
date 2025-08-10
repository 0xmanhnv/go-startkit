package db

import "fmt"

// BuildPostgresDSN constructs a lib/pq DSN string from config values
func BuildPostgresDSN(host, port, user, password, dbname, sslmode string) string {
    return fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        host, port, user, password, dbname, sslmode,
    )
}

