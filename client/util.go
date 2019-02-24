package client

import (
	"crypto/tls"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

func buildHTTPSClient(verify bool) *http.Client {
	return &http.Client{Transport: &http.Transport{
		// Our certificates are self-signed, so will raise
		// a warning - currently, we ask our client to ignore
		// this warning.
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !verify,
		},
	}}
}

func buildWebSocketDialer(verify bool) *websocket.Dialer {
	return &websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !verify,
		},
	}
}

func encodeQuery(url *url.URL, queries map[string]string) {
	q := url.Query()
	for k, v := range queries {
		q.Add(k, v)
	}
	url.RawQuery = q.Encode()
}
