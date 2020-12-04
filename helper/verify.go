package helper

import (
	"regexp"
	"strings"
)

func IsEmpty(str string) bool {
	if str == "" {
		return true
	}
	str = strings.Trim(str, " ")
	if len(str) == 0 {
		return true
	}
	return false
}

func IsEmail(str string) bool {
	pattern := `^([a-zA-Z]|[0-9])(\w|\-)+@[a-zA-Z0-9]+\.([a-zA-Z]{2,4})$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(str)
}

func IsUrl(str string) bool {
	pattern := `^http(s?):\/\/`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(str)
}
