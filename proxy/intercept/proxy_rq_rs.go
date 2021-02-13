package intercept

import (
	"comradequinn/hflow/log"
	"comradequinn/hflow/proxy/internal/codec"
	"comradequinn/hflow/proxy/internal/copy"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ProxyRequest represents a http.ProxyRequest being currently processed by proxy
type ProxyRequest struct {
	URL    url.URL
	Method string
	Header http.Header
	Body   []byte
}

func newProxyRequest(hr *http.Request) (*ProxyRequest, error) {
	r := ProxyRequest{Header: http.Header{}, Method: hr.Method}

	r.URL = *hr.URL

	for k, v := range hr.Header {
		hv := strings.Join(v, " ")
		r.Header.Set(k, hv)
	}

	var err error

	if r.Body, err = copy.CloserToBytes(&hr.Body); err != nil {
		return nil, fmt.Errorf("unable to read request body: [%v]", err)
	}

	return &r, nil
}

func (r *ProxyRequest) http() (*http.Request, error) {
	nr, err := http.NewRequest(r.Method, r.URL.String(), copy.BytesToCloser(r.Body))

	if err == nil {
		copy.Header(r.Header, nr.Header)
	}

	return nr, err
}

// ProxyResponse represents a http.Request being currently processed by proxy
type ProxyResponse struct {
	Header     http.Header
	Body       []byte
	Status     string
	StatusCode int
	Proto      string
	ProtoMajor int
	ProtoMinor int
	Request    *http.Request
	TLS        *tls.ConnectionState
}

func newProxyResponse(hr *http.Response) (*ProxyResponse, error) {
	r := ProxyResponse{Header: http.Header{}}

	copy.Header(hr.Header, r.Header)

	r.Header.Set("Via", "hflow")

	r.Status, r.StatusCode, r.Proto, r.ProtoMajor, r.ProtoMinor, r.Request, r.TLS =
		hr.Status, hr.StatusCode, hr.Proto, hr.ProtoMajor, hr.ProtoMinor, hr.Request, hr.TLS

	var err error

	if r.Body, err = copy.CloserToBytes(&hr.Body); err != nil {
		return nil, fmt.Errorf("unable to read response body: [%v]", err)
	}

	ct := r.Header.Get("Content-Encoding")

	if codec.Supported(ct) {
		log.Printf(1, "decoding response body from [%v] using scheme from content-encoding header [%v]", ct, hr.Request.URL.String())

		if r.Body, err = codec.Decode(ct, r.Body); err != nil {
			return nil, fmt.Errorf("unable to decode response body using scheme from content-encoding header [%v]: [%v]", ct, err)
		}
	}

	return &r, nil
}

func (r *ProxyResponse) http() (*http.Response, error) {
	hr := http.Response{Header: http.Header{}}

	copy.Header(r.Header, hr.Header)

	hr.Status, hr.StatusCode, hr.Proto, hr.ProtoMajor, hr.ProtoMinor, hr.Request, hr.TLS =
		r.Status, r.StatusCode, r.Proto, r.ProtoMajor, r.ProtoMinor, r.Request, r.TLS

	var err error
	ct := r.Header.Get("Content-Encoding")

	if codec.Supported(ct) {
		log.Printf(1, "encoding response body from [%v] using scheme from content-encoding header [%v]", ct, hr.Request.URL.String())

		if r.Body, err = codec.Encode(ct, r.Body); err != nil {
			return nil, fmt.Errorf("unable to encode response body using scheme from content-encoding header [%v]: [%v]", ct, err)
		}
	}

	hr.Body, hr.ContentLength = copy.BytesToCloser(r.Body), int64(len(r.Body))

	return &hr, nil
}
