// copyright @ 2020 ops inc.
//
// author: jinlong yang
//

package g

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

func Get(url string, header map[string]string, timeout int) (string, error) {
	return do(http.MethodGet, url, header, nil, timeout)
}

func Post(url string, header map[string]string, payload []byte, timeout int) (string, error) {
	return do(http.MethodPost, url, header, payload, timeout)
}

func Put(url string, header map[string]string, payload []byte, timeout int) (string, error) {
	return do(http.MethodPut, url, header, payload, timeout)
}

func Patch(url string, header map[string]string, payload []byte, timeout int) (string, error) {
	return do(http.MethodPatch, url, header, payload, timeout)
}

func do(mode, url string, header map[string]string, payload []byte, timeout int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
	defer cancel()

	client := http.Client{Transport: &http.Transport{DisableKeepAlives: true}}

	req, err := http.NewRequest(mode, url, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	if len(header) > 0 {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}
	req = req.WithContext(ctx)
	defer req.Body.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()

	type response struct {
		resp *http.Response
		err  error
	}

	respChan := make(chan *response, 1)
	go func() {
		resp, err := client.Do(req)
		r := &response{resp, err}
		respChan <- r
		wg.Done()
	}()

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("request url: %s timeout", url)
	case r := <-respChan:
		if r.err != nil {
			return "", r.err
		}

		if !Ini(r.resp.StatusCode, []int{200, 201}) {
			return "", fmt.Errorf("request url: %s status code: %d", url, r.resp.StatusCode)
		}

		defer r.resp.Body.Close()
		body, err := ioutil.ReadAll(r.resp.Body)
		if err != nil {
			return "", fmt.Errorf("request url: %s error: %s", url, err)
		}
		return string(body), nil
	}
}
