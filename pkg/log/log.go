package log

import (
	"log"
	"os"
)

var l *log.Logger

func init() {
	l = log.Default()
	f, err := os.OpenFile("", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		l.SetOutput(f)
	}
}

func Info(v ...any) {
	l.Println(v...)
}
