package main

import (
	"fmt"
	"fxtester/internal/common"
	"fxtester/internal/lang"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	_ "github.com/lib/pq"
)

func main() {
	e := echo.New()

	// サービスの初期化
	hdr := NewTestSpService()

	// ミドルウェアの設定
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     common.GetConfig().Server.AllowOrigins,
		AllowCredentials: true,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodDelete},
	}))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "time=${time_rfc3339}, method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.Use(lang.ErrorHandler())

	// サービスの開始
	e.GET("/", hdr.GetHome)
	e.POST("/saml/acs", hdr.PostSamlAcs)
	e.GET("/saml/login", hdr.GetSamlLogin)
	e.GET("/saml/logout", hdr.GetSamlLogout)
	e.POST("/saml/slo", hdr.PostSamlSlo)

	addr := fmt.Sprintf(":%d", TestServerPort)
	sslCertPath := common.GetConfig().Server.Ssl.CertPath
	sslKeyPath := common.GetConfig().Server.Ssl.KeyPath
	e.Logger.Fatal(e.StartTLS(addr, sslCertPath, sslKeyPath))
}
