package net

import (
	"crypto/rand"
	"fxtester/internal/lang"
	"io"
	"time"

	"github.com/golang-jwt/jwt"
)

func makeSecret() []byte {
	rv := make([]byte, 20)

	if _, err := io.ReadFull(rand.Reader, rv); err != nil {
		panic(err)
	}
	return rv
}

var AccessTokenSecret = makeSecret()      // TODO: スケーラブルに問題が出ないように別コマンド化、ファイルに保存？
var RefreshTokenSecret = makeSecret()     // TODO: スケーラブルに問題が出ないように別コマンド化、ファイルに保存？
var SSOSessionSecret = makeSecret()       // TODO: スケーラブルに問題が出ないように別コマンド化、ファイルに保存？
var SLOSessionSecret = makeSecret()       // TODO: スケーラブルに問題が出ないように別コマンド化、ファイルに保存？
var SAMLErrorSessionSecret = makeSecret() // TODO: スケーラブルに問題が出ないように別コマンド化、ファイルに保存？

// クレーム (JWTのペイロード部分)
type Claims[T any] struct {
	Value T `json:"value"`
	jwt.StandardClaims
}

func GenerateToken[T any](value T, expires time.Time, jwtKey []byte) (string, error) {
	claims := &Claims[T]{
		Value: value,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expires.UTC().Unix(),
			IssuedAt:  time.Now().UTC().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", lang.NewFxtError(lang.ErrJWTSign).SetCause(err)
	}

	return tokenString, nil
}

func VerifyToken[T any](tokenStr string, secret []byte) (*Claims[T], error) {
	claims := &Claims[T]{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return nil, lang.NewFxtError(lang.ErrSession).SetCause(err)
	}

	if !token.Valid {
		return nil, lang.NewFxtError(lang.ErrSession)
	}

	return claims, nil
}
