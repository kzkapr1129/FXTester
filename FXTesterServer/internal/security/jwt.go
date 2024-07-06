package security

import (
	"fmt"
	"fxtester/internal/config"
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
			ExpiresAt: expires.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	return generateJWT(claims, config.GetConfig().AccessTokenKey)
}

// GenerateRefreshToken リフレッシュトークンの生成
func GenerateRefreshToken(userId string, expires time.Time) (string, error) {
	claims := &Claims{
		UserId: userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expires.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	return generateJWT(claims, config.GetConfig().RefreshTokenKey)
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
		return config.GetConfig().AccessTokenKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
