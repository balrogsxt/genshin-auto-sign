package helper

import "strings"

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
