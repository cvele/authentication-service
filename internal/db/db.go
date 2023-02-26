package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	"github.com/cvele/authentication-service/internal/config"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type DB struct {
	db *sql.DB
}

type Tx interface {
	InsertUser(id, username, passwordHash string) error
	GetUserByUsername(username string) (*User, error)
	GetUserByID(id string) (*User, error)
	NewUUID() (string, error)
	UpdatePassword(id, password string) error
	Close() error
	Get() *sql.DB
}

type User struct {
	ID       string
	Username string
	Password string
}

func New(cfg *config.Config) (*DB, error) {

	// Connect to the database
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Set connection parameters
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &DB{db}, nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) Get() *sql.DB {
	return db.db
}

func (db *DB) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	err := crdb.ExecuteTx(context.Background(), db.db, nil, func(tx *sql.Tx) error {
		row := tx.QueryRow("SELECT id, username, password FROM users WHERE username = $1", username)
		return row.Scan(&user.ID, &user.Username, &user.Password)
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *DB) InsertUser(id string, username string, password string) error {
	err := crdb.ExecuteTx(context.Background(), db.db, nil, func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO users (id, username, password) VALUES ($1, $2, $3)", id, username, password)
		return err
	})

	return err
}

func (db *DB) NewUUID() (string, error) {
	var id string
	err := crdb.ExecuteTx(context.Background(), db.db, nil, func(tx *sql.Tx) error {
		id = uuid.New().String()
		return nil
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (db *DB) GetUserByID(id string) (*User, error) {
	user := &User{}
	err := crdb.ExecuteTx(context.Background(), db.db, nil, func(tx *sql.Tx) error {
		row := tx.QueryRow("SELECT id, username, password FROM users WHERE id = $1", id)
		return row.Scan(&user.ID, &user.Username, &user.Password)
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (tx *DB) UpdatePassword(id, password string) error {
	err := crdb.ExecuteTx(context.Background(), tx.db, nil, func(tx *sql.Tx) error {
		_, err := tx.Exec("UPDATE users SET password = $1 WHERE id = $2", password, id)
		return err
	})
	return err
}
