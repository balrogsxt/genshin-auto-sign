package helper

import (
	"crypto/md5"
	"encoding/hex"
	jsoniter "github.com/json-iterator/go"
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

//send mail
func SendEmail(targetEmail string, text string) {
	//host := ""
	//user := ""
	//password := ""

	// todo 后续增加发送邮件通知功能
	//intln("send mail....")
	//auth := smtp.PlainAuth("", user, password, host)
	//msg := []byte("测试")
	//
	//to := []string{
	//	"",
	//}
	//
	//err := smtp.SendMail(host, auth, user, to, msg)
	//if err != nil {
	//	log.Info("err->", err)
	//} else {
	//	log.Info("ok")
	//}
}
