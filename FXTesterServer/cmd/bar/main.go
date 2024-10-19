package main

import (
	"fmt"
	"fxtester/internal/common"
	"fxtester/internal/gen"
	"fxtester/internal/lang"
	"fxtester/service"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	_ "github.com/lib/pq"
)

func main() {
	e := echo.New()

	// サービスの初期化
	hdr := service.NewBarService()
	if err := hdr.Init(); err != nil {
		e.Logger.Fatalf("failed to initialize BarService: %v", err)
	}

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
	gen.RegisterHandlers(e, hdr)

	addr := fmt.Sprintf(":%d", common.GetConfig().Server.Port)
	sslCertPath := common.GetConfig().Server.Ssl.CertPath
	sslKeyPath := common.GetConfig().Server.Ssl.KeyPath
	e.Logger.Fatal(e.StartTLS(addr, sslCertPath, sslKeyPath))
}
