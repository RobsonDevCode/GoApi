package configuration

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/gommon/log"
	"time"
)

type DbConfig struct {
	DSN          string
	MaxOpenConns int
	MaxIdelConns int
	MaxLifeTime  time.Duration
	MaxIdelTime  time.Duration
}

func NewDB(cfg DbConfig) (*sql.DB, error) {

	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdelConns)
	db.SetConnMaxLifetime(cfg.MaxLifeTime)
	db.SetConnMaxIdleTime(cfg.MaxIdelTime)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Errorf("failed to ping db %s: %s", cfg.DSN, err)
		return nil, err
	}

	log.Infof("connected to db %s", cfg.DSN)
	return db, nil
}
