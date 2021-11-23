package utils

import (
	"io/ioutil"
	"net/http"
	"net/url"
)

//Get 发送get请求方式
func Get(Url string) ([]byte, error) {
	resp, err := http.Get(Url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

//PostForm 发送post请求方式
func PostForm(Url string, data map[string]string) ([]byte, error) {
	values := url.Values{}
	for k, v := range data {
		values.Add(k, v)
	}

	resp, err := http.PostForm(Url, values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
