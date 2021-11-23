package utils

import (
	"crypto/md5"
	"encoding/hex"
)

//Md5 md5加密
func Md5(s []byte) string {
	h := md5.New()
	h.Write(s)
	return hex.EncodeToString(h.Sum(nil))
}
