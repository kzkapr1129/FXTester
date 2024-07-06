package logger

import (
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

func InitLogger(logFile string, logLevel string) func() {

	closer := func() {}

	// ログ出力先
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		closer = func() { file.Close() }
		logrus.SetOutput(file)
	} else {
		logrus.SetOutput(os.Stdout)
	}

	// ログレベル
	lv, err := logrus.ParseLevel(logLevel)
	if err != nil {
		log.Fatal(err)
	}
	logrus.SetLevel(lv)

	return closer
}
