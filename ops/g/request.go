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

type response struct {
	resp *http.Response
	err  error
}

func Post(url string, header map[string]string, payload []byte, timeout int) (string, error) {
	return do("POST", url, header, payload, timeout)
}

func Put(url string, header map[string]string, payload []byte, timeout int) (string, error) {
	return do("PUT", url, header, payload, timeout)
}

func do(mode, url string, header map[string]string, payload []byte, timeout int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
	defer cancel()

	transport := http.Transport{
		DisableKeepAlives: true,
	}
	client := http.Client{Transport: &transport}

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

	respChan := make(chan *response, 1)
	go func() {
		resp, err := client.Do(req)
		r := &response{resp, err}
		respChan <- r
		wg.Done()
	}()

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("post request url: %s timeout", url)
	case r := <-respChan:
		defer r.resp.Body.Close()
		body, err := ioutil.ReadAll(r.resp.Body)
		if err != nil {
			return "", fmt.Errorf("post request url: %s error: %s", url, err)
		}
		return string(body), nil
	}
}
