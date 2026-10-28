package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Connect initializes a MySQL connection pool, verifies connectivity, and classifies
// common MySQL startup errors into actionable messages without leaking credentials.
func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize mysql driver: %w", err)
	}

	// Connection pool settings suitable for concurrent API requests
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			switch mysqlErr.Number {
			case 1045:
				return nil, fmt.Errorf("authentication failure (MySQL Error 1045): MySQL rejected DB_USER or DB_PASSWORD, or the user does not exist for 'localhost'/'127.0.0.1'")
			case 1044:
				return nil, fmt.Errorf("permission denied (MySQL Error 1044): DB_USER is authenticated but lacks GRANT privileges on DB_NAME")
			case 1049:
				return nil, fmt.Errorf("unknown database (MySQL Error 1049): database specified in DB_NAME does not exist on this MySQL server")
			}
		}
		return nil, fmt.Errorf("unable to reach MySQL server: %w", err)
	}

	return db, nil
}
