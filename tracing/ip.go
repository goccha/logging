package tracing

import (
	"net"
	"net/http"
	"strings"

	"github.com/goccha/http-constants/pkg/headers"
	"github.com/goccha/http-constants/pkg/headers/forwarded"
)

func WithIpHeaders(keys ...IpHeader) Option {
	return func() {
		if len(keys) > 0 {
			_ipHeaders = keys
		}
	}
}

var defaultHeaders = []IpHeader{
	Forwarded(),
	XForwardedFor(),
	RemoteAddr(),
}

func DefaultIpHeaders() IpHeaders {
	return defaultHeaders
}

var _ipHeaders = defaultHeaders

type IpHeader func(req *http.Request) (string, bool)

type IpHeaders []IpHeader

func (h IpHeaders) Get(req *http.Request) (string, bool) {
	for _, header := range h {
		if ip, ok := header(req); ok {
			return ip, true
		}
	}
	return "", false
}
func (h IpHeaders) Prepend(header IpHeader) IpHeaders {
	list := make(IpHeaders, len(h)+1)
	list = append(list, header)
	return append(list, h...)
}
func (h IpHeaders) Append(header IpHeader) IpHeaders {
	list := make(IpHeaders, len(h)+1)
	list = append(list, h...)
	return append(list, header)
}

func Forwarded() IpHeader {
	return func(req *http.Request) (string, bool) {
		if v := req.Header.Get(headers.Forwarded); v != "" {
			return forwarded.Parse(v).ClientIP(), true
		}
		return "", false
	}
}

func XForwardedFor() IpHeader {
	return func(req *http.Request) (string, bool) {
		if ip, ok := getHeaderValue(req, headers.XForwardedFor); ok {
			return ip, true
		}
		return "", false
	}
}

func XRealIp() IpHeader {
	return func(req *http.Request) (string, bool) {
		if ip, ok := getHeaderValue(req, headers.XRealIp); ok {
			return ip, true
		}
		return "", false
	}
}
func RemoteAddr() IpHeader {
	return func(req *http.Request) (string, bool) {
		if ip, _, err := net.SplitHostPort(strings.TrimSpace(req.RemoteAddr)); err != nil && ip != "" {
			return ip, true
		}
		return "", false
	}
}
func XEnvoyExternalAddress() IpHeader {
	return func(req *http.Request) (string, bool) {
		if ip := req.Header.Get(headers.XEnvoyExternalAddress); ip != "" {
			return ip, true
		}
		return "", false
	}
}
func FixedIp(ip string) IpHeader {
	return func(req *http.Request) (string, bool) {
		return ip, true
	}
}
