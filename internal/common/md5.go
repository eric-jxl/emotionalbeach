package common

import (
	"crypto/md5"
	"crypto/sha256"
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

// SaltPassWord 密码加盐 (legacy MD5+salt)
func SaltPassWord(pw string, salt string) string {
	saltPW := fmt.Sprintf("%s$%s", Md5encoder(pw), salt)
	return saltPW
}

// CheckPassWord 核验密码 (legacy MD5+salt)
func CheckPassWord(rpw, salt, pw string) bool {
	return pw == SaltPassWord(rpw, salt)
}

// Sha256Password 使用 sha256 对密码进行哈希（统一前后端加密方式）
func Sha256Password(password string) (string, error) {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:]), nil
}

// CheckSha256Password 验证 sha256 密码
func CheckSha256Password(password, hash string) bool {
	expectedHash, _ := Sha256Password(password)
	return expectedHash == hash
}

// IsSha256Hash 判断存储的密码是否为 sha256 格式
func IsSha256Hash(hash string) bool {
	return len(hash) == 64
}

// IsValidPhoneNumber 定义一个函数来校验手机号
func IsValidPhoneNumber(phone string) bool {
	// 正则表达式匹配中国的手机号
	regex := `^1[3-9]\d{9}$`
	match, _ := regexp.MatchString(regex, phone)
	return match
}
