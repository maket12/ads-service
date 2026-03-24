package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// Registration of driver pgx for using database/sql
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	SSLMode    string

	OpenConn     int
	IdleConn     int
	ConnLifeTime time.Duration
}

func NewConfig(
	host string, port int, user, password, name, ssl string,
	openConn, idleConn int, connLifeTime time.Duration) *Config {
	return &Config{
		DBHost:       host,
		DBPort:       port,
		DBUser:       user,
		DBPassword:   password,
		DBName:       name,
		SSLMode:      ssl,
		OpenConn:     openConn,
		IdleConn:     idleConn,
		ConnLifeTime: connLifeTime,
	}
}

func (pc *Config) dsn() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		pc.DBHost, pc.DBPort, pc.DBUser, pc.DBPassword, pc.DBName, pc.SSLMode,
	)
}

type Client struct {
	DB *sql.DB
}

func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("db config is not specified")
	}

	var dsn = config.dsn()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	db.SetMaxOpenConns(config.OpenConn)
	db.SetMaxIdleConns(config.IdleConn)
	db.SetConnMaxLifetime(config.ConnLifeTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Client{DB: db}, nil
}

func (c *Client) Close() error {
	return c.DB.Close()
}
