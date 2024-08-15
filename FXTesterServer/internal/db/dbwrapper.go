package db

import (
	"database/sql"
	"fxtester/internal/common"
	"fxtester/internal/lang"
	"time"
)

type IDbWrapper interface {
	Init() error
	GetDb() *sql.DB
}

type DbWrapper struct {
	db *sql.DB
}

func (d *DbWrapper) Init() error {
	db, err := sql.Open(common.GetConfig().Db.Name, common.GetConfig().Db.Dsn)
	if err != nil {
		return lang.NewFxtError(lang.ErrDBOpen).SetCause(err)
	}

	// コネクションプールの設定
	db.SetMaxOpenConns(common.GetConfig().Db.MaxOpenConnections)
	db.SetMaxIdleConns(common.GetConfig().Db.MaxIdleConnections)
	db.SetConnMaxLifetime(time.Duration(common.GetConfig().Db.MaxLifeTimeBySec) * time.Second)

	if err := db.Ping(); err != nil {
		return lang.NewFxtError(lang.ErrDBOpen).SetCause(err)
	}

	d.db = db
	return nil
}

func (d *DbWrapper) GetDb() *sql.DB {
	if d.db == nil {
		panic("DB hasn't initialized yet")
	}
	return d.db
}
