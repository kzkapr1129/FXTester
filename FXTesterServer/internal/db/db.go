package db

import (
	"database/sql"
	"fxtester/internal/common"
	"fxtester/internal/lang"
	"time"
)

type IDB interface {
	Init() error
	GetDB() *sql.DB
}

type DB struct {
	db *sql.DB
}

func (d *DB) Init() error {
	db, err := sql.Open(common.GetConfig().DB.Name, common.GetConfig().DB.Dsn)
	if err != nil {
		return lang.NewFxtError(lang.ErrDBOpen).SetCause(err)
	}

	// コネクションプールの設定
	db.SetMaxOpenConns(common.GetConfig().DB.MaxOpenConnections)
	db.SetMaxIdleConns(common.GetConfig().DB.MaxIdleConnections)
	db.SetConnMaxLifetime(time.Duration(common.GetConfig().DB.MaxLifeTimeBySec) * time.Second)

	if err := db.Ping(); err != nil {
		return lang.NewFxtError(lang.ErrDBOpen).SetCause(err)
	}

	d.db = db
	return nil
}

func (d *DB) GetDB() *sql.DB {
	if d.db == nil {
		panic("DB hasn't initialized yet")
	}
	return d.db
}
