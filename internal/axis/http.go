package axis

import (
	"net"
	"net/http"
	"time"
)

// newHTTPClient devuelve un *http.Client con timeouts ajustados para LLM:
// no usamos un timeout global agresivo (Companion puede tardar varios segundos
// en responder) pero sí limitamos handshake/idle para no quedar colgados.
func newHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: timeout,
			IdleConnTimeout:       60 * time.Second,
			MaxIdleConns:          50,
			MaxIdleConnsPerHost:   10,
		},
	}
}
