package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
    http.HandleFunc("/", handleRequestAndRedirect)

    if err := http.ListenAndServe("127.0.0.1:3000", nil); err != nil {
        panic(err)
    }
}

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	url, err := url.Parse("http://127.0.0.1:3001")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ModifyResponse = catchError
	proxy.ErrorHandler = fallback
	proxy.ServeHTTP(res, req)
}

func catchError(res *http.Response) (err error) {
	if res.StatusCode >= 400 {
		return fmt.Errorf("ko")
	}
	return nil
}

func fallback(res http.ResponseWriter, req *http.Request, err error) {
	if err == nil {
		return
	}
	url, err := url.Parse("http://127.0.0.1:3002")
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(res, req)
}
