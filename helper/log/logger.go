package log

import (
	"fmt"
	l "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

var currentTime string = "nil"

type LogFormat struct{}

func (s *LogFormat) Format(entry *l.Entry) ([]byte, error) {
	timestamp := time.Now().Local().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf("[%s] [%s] %s\n", strings.ToUpper(entry.Level.String()), timestamp, entry.Message)
	fmt.Printf(msg)
	return []byte(msg), nil
}
func init() {
}
func logInit() {
	t := time.Now().Format("2006-01-02")
	if currentTime != t {
		currentTime = t
		if !PathExists("logs") {
			if err := os.Mkdir("logs", 0777); err != nil {
				fmt.Printf("[Fatal] 创建日志文件夹失败 \n")
				return
			}
		}
		logFile, err := os.OpenFile(fmt.Sprintf("logs/%s.log", t), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
		if err != nil {
			fmt.Printf("[Fatal] 日志初始化失败: %s \n", err.Error())
			return
		}
		l.SetFormatter(new(LogFormat))
		l.SetOutput(logFile)
		fmt.Printf("[Ok] 日志初始化完毕 \n")
	}

}
func Info(format string, args ...interface{}) {
	logInit()
	l.Info(fmt.Sprintf(format, args...))
}

func Error(format string, args ...interface{}) {
	logInit()
	l.Error(fmt.Sprintf(format, args...))
}
func Warning(format string, args ...interface{}) {
	logInit()
	l.Warning(fmt.Sprintf(format, args...))
}
func Fatal(format string, args ...interface{}) {
	logInit()
	l.Fatal(fmt.Sprintf(format, args...))
}
func Debug(format string, args ...interface{}) {
	logInit()
	l.Debug(fmt.Sprintf(format, args...))
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
