package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"proxy-service/internal/types"
	"strings"

	"golang.org/x/net/proxy"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	return mux
}

func handler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody types.RequestBody
	body, _ := io.ReadAll(req.Body)
	err := json.Unmarshal(body, &reqBody)
	if err != nil || reqBody.Request.URL == "" || reqBody.Proxy.Host == "" || reqBody.Proxy.Port == "" {
		http.Error(w, "400, пошел нахуй, без текста", http.StatusBadRequest)
		return
	}

	method := reqBody.Request.Method
	if method == "" {
		method = http.MethodGet
	}

	var reqUrl *http.Request
	if reqBody.Request.Body != nil && len(reqBody.Request.Body) > 0 {
		reqUrl, _ = http.NewRequest(method, reqBody.Request.URL, strings.NewReader(string(reqBody.Request.Body)))
	} else {
		reqUrl, _ = http.NewRequest(method, reqBody.Request.URL, nil)
	}

	for key, values := range req.Header {
		for _, value := range values {
			reqUrl.Header.Add(key, value)
		}
	}

	client, err := createProxyClient(reqBody)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating SOCKS5 proxy: %v", err), http.StatusInternalServerError)
		return
	}

	resUrl, err := client.Do(reqUrl)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching URL: %v", err), http.StatusInternalServerError)
		return
	}
	defer resUrl.Body.Close()

	for key, values := range resUrl.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resUrl.StatusCode)

	buffer := make([]byte, 4096)
	for {
		n, err := resUrl.Body.Read(buffer)
		if err != nil && err != io.EOF {
			http.Error(w, fmt.Sprintf("Error Reading response body: %v\n", err), http.StatusInternalServerError)
			return
		}

		if n > 0 {
			_, err := w.Write(buffer[:n])
			if err != nil {
				http.Error(w, fmt.Sprintf("Error writing response: %v\n", err), http.StatusInternalServerError)
				return
			}
			w.(http.Flusher).Flush()
		}

		if err == io.EOF {
			break
		}
	}
}

func createProxyClient(body types.RequestBody) (*http.Client, error) {
	client := &http.Client{}

	auth := proxy.Auth{User: body.Proxy.Login, Password: body.Proxy.Password}
	proxyAddr := fmt.Sprintf("%s:%s", body.Proxy.Host, body.Proxy.Port)
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, &auth, proxy.Direct)

	if err != nil {
		return nil, err
	}

	client = &http.Client{
		Transport: &http.Transport{Dial: dialer.Dial},
	}

	return client, nil
}
