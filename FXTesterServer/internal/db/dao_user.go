package db

import (
	"database/sql"
	"errors"
	"fxtester/internal/lang"
)

type Token struct {
	AccessToken  string
	RefreshToken string
}

type IDaoBase interface {
	Begin() error
	Rollback() error
	Commit() error
	Query(query string, args ...any) (*sql.Rows, error)
}

type DaoBase struct {
	dbWrapper IDbWrapper
	tx        *sql.Tx
}

func (e *DaoBase) Begin() error {
	db := e.dbWrapper.GetDb()
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
	return e.dbWrapper.GetDb().Query(query, args...)
}

type IUserEntityDao interface {
	IDaoBase
	CreateUser(email string) (*UserEntity, error)
	UpdateToken(userId int64, accessToken, refreshToken string) error
	CheckAccessToken(userId int64, accessToken string) error
	CheckRefreshToken(userId int64, refreshToken string) error
	SelectWithUserId(userId int64) (*UserEntity, error)
	SelectWithEmail(email string) (*UserEntity, error)
}

type UserEntityDao struct {
	DaoBase
}

func NewUserEntityDao(dbWrapper IDbWrapper) IUserEntityDao {
	return &UserEntityDao{
		DaoBase: DaoBase{
			dbWrapper: dbWrapper,
		},
	}
}

// CreateUser userテーブルに新規レコードを追加する
func (u *UserEntityDao) CreateUser(email string) (user *UserEntity, lastError error) {
	rows, err := u.DaoBase.Query("select fxtester_schema.create_user($1)", email)
	if err != nil {
		return nil, lang.NewFxtError(lang.ErrDBQuery).SetCause(err)
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, lang.NewFxtError(lang.ErrDBQueryResult).SetCause(err)
	}
	var newId int64
	if err := rows.Scan(&newId); err != nil {
		return nil, lang.NewFxtError(lang.ErrDBQueryResult).SetCause(err)
	}

	return &UserEntity{
		UserId: newId,
		Email:  email,
	}, nil
}

func (u *UserEntityDao) UpdateToken(userId int64, accessToken, refreshToken string) error {
	rows, err := u.DaoBase.Query("call fxtester_schema.update_token($1, $2, $3)", userId, accessToken, refreshToken)
	if err != nil {
		return nil
	}
	defer rows.Close()
	return nil
}

func (u *UserEntityDao) CheckAccessToken(userId int64, accessToken string) error {
	return errors.New("not implements")
}

func (u *UserEntityDao) CheckRefreshToken(userId int64, refreshToken string) error {
	return errors.New("not implements")
}

func (u *UserEntityDao) SelectWithUserId(userId int64) (*UserEntity, error) {
	return nil, errors.New("not implements")
}

func (u *UserEntityDao) SelectWithEmail(email string) (*UserEntity, error) {
	sql := `
		select
			id,
			email,
			access_token,
			refresh_token
		from fxtester_schema.select_user_with_email($1)
	`
	rows, err := u.DaoBase.Query(sql, email)
	if err != nil {
		return nil, lang.NewFxtError(lang.ErrDBQuery).SetCause(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, lang.NewFxtError(lang.ErrDBQueryResult).SetCause(err)
	}

	var user UserEntity
	if err := rows.Scan(&user.UserId, &user.Email, &user.AccessToken, &user.RefreshToken); err != nil {
		return nil, lang.NewFxtError(lang.ErrDBQueryResult).SetCause(err)
	}

	return &user, nil
}
