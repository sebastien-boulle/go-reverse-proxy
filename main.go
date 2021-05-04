package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	listen = "127.0.0.1:3000"
	serviceUrl = "https://service.local"
	fallbackUrl = "https://fallback.local"
)

func main() {
    http.HandleFunc("/", handleRequestAndRedirect)

    if err := http.ListenAndServe(listen, nil); err != nil {
        panic(err)
    }
}

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	url, err := url.Parse(serviceUrl)
	if err != nil {
		panic(err)
	}
	body, _ := ioutil.ReadAll(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewReader(body))

	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ModifyResponse = catchError
	proxy.ErrorHandler = fallback(body)
	proxy.ServeHTTP(res, req)
}

func catchError(res *http.Response) (err error) {
	if res.StatusCode >= 400 {
		return fmt.Errorf("ko")
	}
	return nil
}

func fallback(body []byte) func(res http.ResponseWriter, req *http.Request, err error) {
	localBody := body
	return func(res http.ResponseWriter, req *http.Request, err error) {
		if err == nil {
			return
		}
		url, err := url.Parse(fallbackUrl)
		if err != nil {
			panic(err)
		}
		proxy := httputil.NewSingleHostReverseProxy(url)

		req.URL.Host = url.Host
		req.URL.Scheme = url.Scheme
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.Host = url.Host
		req.Body = ioutil.NopCloser(bytes.NewReader(localBody))
		proxy.ServeHTTP(res, req)
	}
}
