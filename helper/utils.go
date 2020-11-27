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
