package db

import (
	"errors"
	"fxtester/internal/lang"
)

type Token struct {
	AccessToken  string
	RefreshToken string
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
	IDaoBase
}

func NewUserEntityDao(dbWrapper IDbWrapper) IUserEntityDao {
	return &UserEntityDao{
		IDaoBase: &DaoBase{
			dbWrapper: dbWrapper,
		},
	}
}

// CreateUser userテーブルに新規レコードを追加する
func (u *UserEntityDao) CreateUser(email string) (user *UserEntity, lastError error) {
	rows, err := u.IDaoBase.Query("select fxtester_schema.create_user($1)", email)
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
	rows, err := u.IDaoBase.Query("call fxtester_schema.update_token($1, $2, $3)", userId, accessToken, refreshToken)
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
	rows, err := u.IDaoBase.Query(sql, email)
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
