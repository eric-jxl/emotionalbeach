package common

import "regexp"

// IsValidPhoneNumber 定义一个函数来校验手机号
func IsValidPhoneNumber(phone string) bool {
	// 正则表达式匹配中国的手机号
	regex := `^1[3-9]\d{9}$`
	match, _ := regexp.MatchString(regex, phone)
	return match
}
