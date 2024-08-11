package internal

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

// クレーム (JWTのペイロード部分)
type Claims struct {
	UserId string `json:"user_id"`
	jwt.StandardClaims
}

// GenerateAccessToken アクセストークンの生成
func GenerateAccessToken(userId string, expires time.Time) (string, error) {
	claims := &Claims{
		UserId: userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expires.UTC().Unix(),
			IssuedAt:  time.Now().UTC().Unix(),
		},
	}
	return generateJWT(claims, GetConfig().Server.JwtKey.AccessToken)
}

// GenerateRefreshToken リフレッシュトークンの生成
func GenerateRefreshToken(userId string, expires time.Time) (string, error) {
	claims := &Claims{
		UserId: userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expires.UTC().Unix(),
			IssuedAt:  time.Now().UTC().Unix(),
		},
	}
	return generateJWT(claims, GetConfig().Server.JwtKey.RefreshToken)
}

func generateJWT(claims jwt.Claims, jwtKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyAccessToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return GetConfig().Server.JwtKey.AccessToken, nil
	})

	if err != nil {
		fmt.Println("failed ParseWithClaims: ", err)
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
