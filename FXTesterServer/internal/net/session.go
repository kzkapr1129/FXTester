package net

import (
	"fxtester/internal/gen"
	"fxtester/internal/lang"
	"net/http"
	"time"
)

const (
	NameAccessToken    = "access_token"
	NameRefreshToken   = "refresh_token"
	NameSSOToken       = "sso_token"
	NameSLOToken       = "slo_token"
	NameSAMLErrorToken = "saml_error_token"
)

type AuthSessionPayload struct {
	UserId int64  `json:"user_id"`
	Email  string `json:"email"`
}

func CreateAuthSession(w http.ResponseWriter, userId int64, email string, onNewToken func(accessToken, refreshToken string) error) error {
	now := time.Now()
	expiresAccessToken := now.Add(15 * time.Minute)
	expiresRefreshToken := now.Add(7 * 24 * time.Hour)

	payload := AuthSessionPayload{
		UserId: userId,
		Email:  email,
	}

	accessToken, err := GenerateToken(payload, expiresAccessToken, AccessTokenSecret)
	if err != nil {
		return err
	}

	refreshToken, err := GenerateToken(payload, expiresRefreshToken, RefreshTokenSecret)
	if err != nil {
		return err
	}

	if err := onNewToken(accessToken, refreshToken); err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     NameAccessToken,
		Value:    accessToken,
		Expires:  expiresAccessToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     NameRefreshToken,
		Value:    refreshToken,
		Expires:  expiresRefreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/", // TODO PATHを限定する
	})

	return nil
}

func GetAuthSessionAccessToken(r *http.Request) (*AuthSessionPayload, error) {
	cookie, err := r.Cookie(NameAccessToken)
	if err != nil {
		return nil, lang.NewFxtError(lang.ErrCookieNone).SetCause(err)
	}
	token := cookie.Value
	claims, err := VerifyToken[AuthSessionPayload](token, AccessTokenSecret)
	if err != nil {
		return nil, err
	}
	return &claims.Value, nil
}

func DeleteAuthSession(w http.ResponseWriter) {

	http.SetCookie(w, &http.Cookie{
		Name:     NameAccessToken,
		Value:    "",
		Expires:  time.Unix(0, 0), // 過去の日付を設定してCookieを削除する
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/api",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     NameRefreshToken,
		Value:    "",
		Expires:  time.Unix(0, 0), // 過去の日付を設定してCookieを削除する
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/api/auth/refresh",
	})
}

type SSOSessionPayload struct {
	AuthnRequestId     string `json:"authnRequestId"`
	RedirectURL        string `json:"redirectURL"`
	RedirectURLOnError string `json:"redirectURLOnError"`
}

func CreateSSOSession(w http.ResponseWriter, authnRequestId string, redirectURL string, redirectURLOnError string) error {
	now := time.Now()
	expires := now.Add(60 * time.Minute)

	payload := SSOSessionPayload{
		AuthnRequestId:     authnRequestId,
		RedirectURL:        redirectURL,
		RedirectURLOnError: redirectURLOnError,
	}

	token, err := GenerateToken(payload, expires, SSOSessionSecret)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     NameSSOToken,
		Value:    token,
		Expires:  expires,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/saml/acs",
	})

	return nil
}

func GetSSOSession(r *http.Request) (*SSOSessionPayload, error) {
	cookie, err := r.Cookie(NameSSOToken)
	if err != nil {
		return nil, lang.NewFxtError(lang.ErrCookieNone).SetCause(err)
	}
	token := cookie.Value
	claims, err := VerifyToken[SSOSessionPayload](token, SSOSessionSecret)
	if err != nil {
		return nil, err
	}
	return &claims.Value, nil
}

func DeleteSSOSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     NameSSOToken,
		Value:    "",
		Expires:  time.Unix(0, 0), // 過去の日付を設定してCookieを削除する
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/saml/acs",
	})
}

type SLOSessionPayload struct {
	UserId             int64  `json:"userId"`
	AuthnRequestId     string `json:"authnRequestId"`
	RedirectURL        string `json:"redirectURL"`
	RedirectURLOnError string `json:"redirectURLOnError"`
}

func CreateSLOSession(w http.ResponseWriter, userId int64, authnRequestId string, redirectURL string, redirectURLOnError string) error {
	now := time.Now()
	expires := now.Add(60 * time.Minute)

	payload := SLOSessionPayload{
		UserId:             userId,
		AuthnRequestId:     authnRequestId,
		RedirectURL:        redirectURL,
		RedirectURLOnError: redirectURLOnError,
	}

	token, err := GenerateToken(payload, expires, SLOSessionSecret)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     NameSLOToken,
		Value:    token,
		Expires:  expires,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/saml/slo",
	})

	return nil
}

func GetSLOSession(r *http.Request) (*SLOSessionPayload, error) {
	cookie, err := r.Cookie(NameSLOToken)
	if err != nil {
		return nil, lang.NewFxtError(lang.ErrCookieNone).SetCause(err)
	}
	token := cookie.Value
	claims, err := VerifyToken[SLOSessionPayload](token, SLOSessionSecret)
	if err != nil {
		return nil, err
	}
	return &claims.Value, nil
}

func DeleteSLOSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     NameSLOToken,
		Value:    "",
		Expires:  time.Unix(0, 0), // 過去の日付を設定してCookieを削除する
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/saml/slo",
	})
}

func CreateSamlErrorSession(w http.ResponseWriter, genErr gen.Error) error {
	now := time.Now()
	expires := now.Add(5 * time.Minute)

	payload := gen.ErrorWithTime{
		Err:  genErr,
		Time: time.Now().Format(time.RFC3339),
	}

	token, err := GenerateToken(payload, expires, SAMLErrorSessionSecret)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     NameSAMLErrorToken,
		Value:    token,
		Expires:  expires,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/saml/error",
	})

	return nil
}

func GetSamlErrorSession(r *http.Request) (*gen.ErrorWithTime, error) {
	cookie, err := r.Cookie(NameSAMLErrorToken)
	if err != nil {
		return nil, lang.NewFxtError(lang.ErrCookieNone).SetCause(err)
	}
	token := cookie.Value
	claims, err := VerifyToken[gen.ErrorWithTime](token, SAMLErrorSessionSecret)
	if err != nil {
		return nil, err
	}
	return &claims.Value, nil
}

func DeleteSamlErrorSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     NameSAMLErrorToken,
		Value:    "",
		Expires:  time.Unix(0, 0), // 過去の日付を設定してCookieを削除する
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Path:     "/saml/error",
	})
}
