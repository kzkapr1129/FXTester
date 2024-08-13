package db

import (
	"database/sql"
	"fxtester/internal/common"
	"fxtester/internal/lang"
	"time"
)

type IProvider interface {
	Init() error
	GetHandle() *sql.DB
}

type Provider struct {
	db *sql.DB
}

func (d *Provider) Init() error {
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

func (d *Provider) GetHandle() *sql.DB {
	if d.db == nil {
		panic("DB hasn't initialized yet")
	}
	return d.db
}
