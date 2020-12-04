package log

import (
	"fmt"
	"time"
)

func Info(format string, args ...interface{}) {
	if len(format) == 0 {
		fmt.Println()
	} else {
		t := time.Now().Format("2006-01-02 15:04:05")
		fmt.Println(fmt.Sprintf("["+t+"] "+format, args...))
	}
}
