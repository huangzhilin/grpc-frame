package utils

import (
	"net/url"
)

// URLUtil url转换
type URLUtil struct {
	rawString   string
	queryObject url.Values
	fragment    string
	host        string
	pathname    string
	scheme      string
}

// ParseURL 解析url string
func ParseURL(raw string) (*URLUtil, error) {
	var urlObject, err = url.Parse(raw)
	if err != nil {
		return nil, err
	}
	var u = &URLUtil{}
	u.rawString = raw
	u.scheme = urlObject.Scheme
	u.host = urlObject.Host
	u.pathname = urlObject.Path
	u.queryObject = urlObject.Query()
	u.fragment = urlObject.Fragment
	return u, nil
}

// SetParam 设置参数
func (u *URLUtil) SetParam(key, value string) {
	u.queryObject.Set(key, value)
}

// Encode 编码生成url
func (u *URLUtil) Encode() string {
	return u.scheme + "://" + u.host + u.pathname + "?" + u.queryObject.Encode() + "#" + u.fragment
}

// Query 获取参数
func (u *URLUtil) Query() url.Values {
	return u.queryObject
}
