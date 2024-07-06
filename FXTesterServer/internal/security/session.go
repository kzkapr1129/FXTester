package security

import (
	"net/http"
	"time"
)

const (
	NAME_ACCESS_TOKEN  = "access_token"
	NAME_REFRESH_TOKEN = "refresh_token"
)

func CreateSession(w http.ResponseWriter, userId string) (string, error) {
	now := time.Now()
	expiresAccessToken := now.Add(15 * time.Minute)
	expiresRefreshToken := now.Add(7 * 24 * time.Hour)

	accessToken, err := GenerateAccessToken(userId, expiresAccessToken)
	if err != nil {
		return "", err
	}

	refreshToken, err := GenerateRefreshToken(userId, expiresRefreshToken)
	if err != nil {
		return "", err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     NAME_ACCESS_TOKEN,
		Value:    accessToken,
		Expires:  expiresAccessToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/api",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     NAME_REFRESH_TOKEN,
		Value:    refreshToken,
		Expires:  expiresRefreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/api/auth/refresh",
	})

	return expiresAccessToken.UTC().Format(time.RFC3339), nil
}

func DeleteSession(w http.ResponseWriter) {

	http.SetCookie(w, &http.Cookie{
		Name:     NAME_ACCESS_TOKEN,
		Value:    "",
		Expires:  time.Unix(0, 0), // 過去の日付を設定してCookieを削除する
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/api",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     NAME_REFRESH_TOKEN,
		Value:    "",
		Expires:  time.Unix(0, 0), // 過去の日付を設定してCookieを削除する
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/api/auth/refresh",
	})
}

func GetAccessToken(r http.Request) (string, error) {
	cookie, err := r.Cookie(NAME_ACCESS_TOKEN)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func GetRefreshToken(r http.Request) (string, error) {
	cookie, err := r.Cookie(NAME_REFRESH_TOKEN)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
