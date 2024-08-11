package main

import (
	"fmt"
	"fxtester/internal"
	"fxtester/internal/gen"
	"fxtester/service"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	// サービスの初期化
	hdr, err := service.NewBarService()
	if err != nil {
		e.Logger.Fatalf("failed to load config: %v", err)
	}

	// ミドルウェアの設定
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     internal.GetConfig().Server.AllowOrigins,
		AllowCredentials: true,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodDelete},
	}))
	e.Use(internal.ErrorHandler())

	// サービスの開始
	gen.RegisterHandlers(e, hdr)

	addr := fmt.Sprintf(":%d", internal.GetConfig().Server.Port)
	sslCertPath := internal.GetConfig().Server.Ssl.CertPath
	sslKeyPath := internal.GetConfig().Server.Ssl.KeyPath
	e.Logger.Fatal(e.StartTLS(addr, sslCertPath, sslKeyPath))
}
