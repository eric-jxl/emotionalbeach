package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
)

// Md5encoder 加密后返回小写值
func Md5encoder(code string) string {
	m := md5.New()
	_, _ = io.WriteString(m, code)
	return hex.EncodeToString(m.Sum(nil))
}

// SaltPassWord 密码加盐
func SaltPassWord(pw string, salt string) string {
	saltPW := fmt.Sprintf("%s$%s", Md5encoder(pw), salt)
	return saltPW
}

// CheckPassWord 核验密码
func CheckPassWord(rpw, salt, pw string) bool {
	return pw == SaltPassWord(rpw, salt)
}

// IsValidPhoneNumber 定义一个函数来校验手机号
func IsValidPhoneNumber(phone string) bool {
	// 正则表达式匹配中国的手机号
	regex := `^1[3-9]\d{9}$`
	match, _ := regexp.MatchString(regex, phone)
	return match
}
