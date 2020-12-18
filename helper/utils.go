package helper

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/gomail.v2"
)

var jsonConfig = jsoniter.ConfigCompatibleWithStandardLibrary

//任意结构体转json字符串
func JsonEncode(v interface{}) string {
	s, _ := jsonConfig.Marshal(&v)
	return string(s)
}

//字符串转结构体
func JsonDecode(jsonString string, v interface{}) error {
	return jsonConfig.Unmarshal([]byte(jsonString), &v)
}
func Md5(str string) string {
	m := md5.New()
	m.Write([]byte(str))
	return hex.EncodeToString(m.Sum(nil))
}

func SendEmail(to string, title string, body string) error {
	conf := GetConfig().Smtp
	if conf.Enable {
		m := gomail.NewMessage(gomail.SetCharset("utf-8"))
		m.SetHeader("From", m.FormatAddress(conf.User, conf.From))
		m.SetHeader("To", to)
		m.SetHeader("Subject", title)
		m.SetBody("text/html", body)
		d := gomail.NewDialer(conf.Host, conf.Port, conf.User, conf.Password)
		err := d.DialAndSend(m)
		return err
	} else {
		return errors.New("当前服务端未启用邮件发送")
	}

}
