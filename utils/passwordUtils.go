package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// EncryptPassword 对密码进行md5加密
func EncryptPassword(password string) string {
	hash := md5.New()
	hash.Write([]byte(password))
	password = hex.EncodeToString(hash.Sum(nil))
	return password
}
