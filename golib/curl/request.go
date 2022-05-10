// copyright @ 2020 ops inc.
//
// author: jinlong yang
//

package curl

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Get 请求 目前需要将参数拼接到url后面
func Get(url string, header map[string]string, timeout int) (string, error) {
	_, _, body, err := Curl(http.MethodGet, url, nil, header, nil, timeout)
	return body, err
}

// Post 请求 线上尽量全部使用post请求
func Post(url string, header map[string]string, payload []byte, timeout int) (string, error) {
	_, _, body, err := Curl(http.MethodPost, url, nil, header, payload, timeout)
	return body, err
}

func Put(url string, header map[string]string, payload []byte, timeout int) (string, error) {
	_, _, body, err := Curl(http.MethodPut, url, nil, header, payload, timeout)
	return body, err
}

func Patch(url string, header map[string]string, payload []byte, timeout int) (string, error) {
	_, _, body, err := Curl(http.MethodPatch, url, nil, header, payload, timeout)
	return body, err
}

func Curl(mode, url string, cookie map[string]string, header map[string]string, payload []byte, timeout int) (map[string]string, map[string]string, string, error) {
	request, err := http.NewRequest(mode, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, nil, "", err
	}
	if len(cookie) > 0 {
		for k, v := range cookie {
			ck := http.Cookie{Name: k, Value: v}
			request.AddCookie(&ck)
		}
	}
	if len(header) > 0 {
		for k, v := range header {
			request.Header.Set(k, v)
		}
	}

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
	defer cancel()
	request = request.WithContext(ctx)

	type result struct {
		response *http.Response
		err      error
	}
	resultChan := make(chan *result, 1)

	go func() {
		client := http.Client{Transport: &http.Transport{DisableKeepAlives: true}}
		response, err := client.Do(request)
		resultChan <- &result{response, err}
	}()

	select {
	case <-ctx.Done():
		return nil, nil, "", fmt.Errorf("request url: %s timeout", url)
	case r := <-resultChan:
		if r.err != nil {
			return nil, nil, "", r.err
		}

		rc, rh, body, err := Parse(r.response, url)
		if err != nil {
			return rc, rh, "", err
		}
		return rc, rh, body, nil
	}
}

func Parse(response *http.Response, url string) (map[string]string, map[string]string, string, error) {
	statusCode := response.StatusCode
	if statusCode != 200 && statusCode != 201 {
		return nil, nil, "", fmt.Errorf("request url: %s status code: %d", url, statusCode)
	}

	// 获取cookie
	cookie := make(map[string]string)
	for _, item := range response.Cookies() {
		cookie[item.Name] = item.Value
	}

	// 获取header
	header := make(map[string]string)
	for k, vs := range response.Header {
		header[k] = strings.Join(vs, "")
	}

	// 获取body
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return cookie, header, "", fmt.Errorf("read url: %s body error: %s", url, err)
	}
	return cookie, header, string(body), nil
}
