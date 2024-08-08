package types

import "encoding/json"

type RequestBody struct {
	Proxy   ProxyConfig   `json:"proxy"`
	Request RequestConfig `json:"request"`
}

type ProxyConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
}

type RequestConfig struct {
	Method string          `json:"method,omitempty"`
	URL    string          `json:"url"`
	Body   json.RawMessage `json:"body,omitempty"`
}
