package db

import (
	"database/sql"
	"fxtester/internal/lang"
)

type IDaoBase interface {
	Begin() error
	Rollback() error
	Commit() error
	Query(query string, args ...any) (*sql.Rows, error)
}

type DaoBase struct {
	db IDB
	tx *sql.Tx
}

func (e *DaoBase) Begin() error {
	db := e.db.GetDB()
	tx, err := db.Begin()
	if err != nil {
		return lang.NewFxtError(lang.ErrDBBegin).SetCause(err)
	}
	e.tx = tx
	return nil
}

func (e *DaoBase) Rollback() error {
	if e.tx != nil {
		defer func() {
			e.tx = nil
		}()
		err := e.tx.Rollback()
		if err != nil {
			return lang.NewFxtError(lang.ErrDBRollback).SetCause(err)
		}
	}
	return nil
}

func (e *DaoBase) Commit() error {
	if e.tx != nil {
		defer func() {
			e.tx = nil
		}()
		err := e.tx.Commit()
		if err != nil {
			return lang.NewFxtError(lang.ErrDBCommit).SetCause(err)
		}
	}
	return nil
}

func (e *DaoBase) Query(query string, args ...any) (*sql.Rows, error) {
	if e.tx != nil {
		return e.tx.Query(query, args...)
	}
	return e.db.GetDB().Query(query, args...)
}
