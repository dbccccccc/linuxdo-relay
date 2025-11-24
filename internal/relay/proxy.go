package relay

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

// ProxyClient is a thin HTTP client wrapper used to forward requests to new-api.
type ProxyClient struct {
	HTTP *http.Client
}

func NewProxyClient() *ProxyClient {
	return &ProxyClient{HTTP: &http.Client{}}
}

// ProxyRequest forwards the given body to the target URL with the provided
// method and headers. It copies the upstream response back to w without
// inspecting or modifying it.
//
// It returns the upstream HTTP status code (if the request was sent
// successfully) and any error encountered while performing the request or
// copying the response body.
func (c *ProxyClient) ProxyRequest(w http.ResponseWriter, origReq *http.Request, method, url, apiKey string, body []byte) (int, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	upReq, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return 0, err
	}

	// Copy relevant headers from original request.
	for k, vals := range origReq.Header {
		// Skip original Authorization; we will set channel-specific key below.
		if strings.EqualFold(k, "Authorization") {
			continue
		}
		for _, v := range vals {
			upReq.Header.Add(k, v)
		}
	}
	if apiKey != "" {
		upReq.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := c.HTTP.Do(upReq)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Copy status and headers.
	for k, vals := range resp.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	return resp.StatusCode, err
}
