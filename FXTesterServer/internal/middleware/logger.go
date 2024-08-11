package middleware

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type ctxKey string

const (
	logKey = ctxKey("logger")
)

func NewLogger(log *logrus.Logger) func(next echo.HandlerFunc) echo.HandlerFunc {
	// 新しいインスタンスを生成しておく
	logger := logrus.NewEntry(log)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			// Contextにlogrusを格納する
			req := c.Request()
			ctx := req.Context()
			ctx = context.WithValue(ctx, logKey, logger)
			req = req.WithContext(ctx)
			c.SetRequest(req)

			// 処理時間の計測
			start := time.Now()
			err := next(c)
			stop := time.Now()
			latency := stop.Sub(start)

			// ログ出力
			log.WithFields(logrus.Fields{
				"method":     c.Request().Method,
				"uri":        c.Request().RequestURI,
				"status":     c.Response().Status,
				"latency":    latency,
				"remote_ip":  c.RealIP(),
				"host":       c.Request().Host,
				"user_agent": c.Request().UserAgent(),
				"error":      err,
			}).Debug("handled request")

			return err
		}
	}
}

func GetLogger(c echo.Context) *logrus.Entry {
	if v, ok := c.Request().Context().Value(logKey).(*logrus.Entry); ok {
		return v
	} else {
		return logrus.NewEntry(logrus.StandardLogger())
	}
}
