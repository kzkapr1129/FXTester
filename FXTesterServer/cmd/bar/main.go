package main

import (
	"flag"
	"fmt"
	"fxtester/internal"
	fxtm "fxtester/middleware"
	"fxtester/openapi/gen"
	"fxtester/service"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

func main() {
	logLevel := flag.String("log-level", "debug", "ログレベル")
	logfile := flag.String("log-out", "", "ログ出力先ファイル名")
	flag.Parse()

	// ログの初期化
	closer := internal.InitLogger(*logfile, *logLevel)
	defer closer()

	e := echo.New()

	// サービスの初期化
	hdr, err := service.NewBarService()
	if err != nil {
		e.Logger.Fatalf("failed to load config: %v", err)
	}

	// ミドルウェアの設定
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     internal.GetConfig().AllowOrigins,
		AllowCredentials: true,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodDelete},
	}))
	e.Use(fxtm.NewLogger(logrus.StandardLogger()))

	// サービスの開始
	gen.RegisterHandlers(e, hdr)

	addr := fmt.Sprintf(":%d", internal.GetConfig().Port)
	sslCertPath := internal.GetConfig().SslCertPath
	sslKeyPath := internal.GetConfig().SslKeyPath
	e.Logger.Fatal(e.StartTLS(addr, sslCertPath, sslKeyPath))
}
