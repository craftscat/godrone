package utils

import (
	"io"
	"log"
	"os"
)

func LoggingSetting(logFile string) {
	// ログファイルを読み書き可能な形で開く
	logfile, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("file=logFile err=%s", err.Error())
	}
	// 実行時とログファイルに書き込む
	multiLogFile := io.MultiWriter(os.Stdout, logfile)
	// 書き込み設定
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(multiLogFile)
}
