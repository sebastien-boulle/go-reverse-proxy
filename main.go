package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"time"
	"fmt"
	"net/http"
)

const (
        proxyPort   = 3000
        servicePort = 3001
        fallbackPort = 3002
)

type Proxy struct{}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Forward the HTTP request to the destination service.
	body, _ := ioutil.ReadAll(req.Body)

        res, duration, err := p.forwardRequest(req)
	if err == nil {
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
	        p.printStats(req, res, duration)
	}

	// Notify the client if there was an error while forwarding the request.
        if err != nil || res.StatusCode >= 400 {
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
		res, duration, err = p.fallbackRequest(req)
        }

        if err != nil {
                http.Error(w, err.Error(), http.StatusBadGateway)
                return
        }
	p.printStats(req, res, duration)

	// If the request was forwarded successfully, write the response back to
	// the client.
        p.writeResponse(w, res)
}

func main() {
	// Listen on the predefined proxy port.
        http.ListenAndServe(fmt.Sprintf(":%d", proxyPort), &Proxy{})
}

func (p *Proxy) forwardRequest(req *http.Request) (*http.Response, time.Duration, error) {
	// Prepare the destination endpoint to forward the request to.
        proxyUrl := fmt.Sprintf("http://127.0.0.1:%d%s", servicePort, req.RequestURI)

	// Print the original URL and the proxied request URL.
        fmt.Printf("Original URL: http://%s%s\n", req.Host, req.RequestURI)
        fmt.Printf("Proxy URL: %s\n", proxyUrl)

	// Create an HTTP client and a proxy request based on the original request.
        httpClient := http.Client{}
        proxyReq, err := http.NewRequest(req.Method, proxyUrl, req.Body)

	// Capture the duration while making a request to the destination service.
        start := time.Now()
        res, err := httpClient.Do(proxyReq)
        duration := time.Since(start)

	// Return the response, the request duration, and the error.
        return res, duration, err
}

func (p *Proxy) fallbackRequest(req *http.Request) (*http.Response, time.Duration, error) {
	// Prepare the destination endpoint to forward the request to.
        proxyUrl := fmt.Sprintf("http://127.0.0.1:%d%s", fallbackPort, req.RequestURI)

	// Print the original URL and the proxied request URL.
        fmt.Printf("Original URL: http://%s%s\n", req.Host, req.RequestURI)
        fmt.Printf("Proxy URL: %s\n", proxyUrl)

	// Create an HTTP client and a proxy request based on the original request.
        httpClient := http.Client{}
        proxyReq, err := http.NewRequest(req.Method, proxyUrl, req.Body)

	// Capture the duration while making a request to the destination service.
        start := time.Now()
        res, err := httpClient.Do(proxyReq)
        duration := time.Since(start)

	// Return the response, the request duration, and the error.
        return res, duration, err
}

func (p *Proxy) writeResponse(w http.ResponseWriter, res *http.Response) {
	// Copy all the header values from the response.
        for name, values := range res.Header {
                w.Header()[name] = values
        }

        w.Header().Set("Server", "amazing-proxy")

	// Set the status code returned by the destination service.
        w.WriteHeader(res.StatusCode)

	// Copy the contents from the response body.
        io.Copy(w, res.Body)

	// Finish the request.
        res.Body.Close()
}

func (p *Proxy) printStats(req *http.Request, res *http.Response, duration time.Duration) {
        fmt.Printf("Request Duration: %v\n", duration)
        fmt.Printf("Request Size: %d\n", req.ContentLength)
        fmt.Printf("Response Size: %d\n", res.ContentLength)
        fmt.Printf("Response Status: %d\n\n", res.StatusCode)
}
