package proxy

import (
	"comradequinn/hflow/proxy/intercept"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestProxy(t *testing.T) {
	test := func(t *testing.T, icpt bool, clientTLS *tls.Config, proxyHandler http.HandlerFunc, newStubSvrFunc func(http.Handler) *httptest.Server) {
		var rcvMethod, rcvPath, rcvQSV, rcvHdrV, rcvBdy, rcvIntHdrV string

		rsCode, rsHdrK, rsHdrV, rsBdy := http.StatusOK, "rsqHdrK", "rsHdrV", "rsBdy"

		rqMethod, rqPath, rqQSK, rqQSV, rqHdrK, rqHdrV, rqBdy, intRqHdrK, intRqHdrV, intRsHdrK, intRsHdrV :=
			http.MethodPost, "test-path", "rqQSK", "rqQSV", "rqHdrK", "rqHdrV", "rqBdy", "intRqHdrK", "intRqHdrV", "intRsHdrK", "intRsHdrV"

		proxy, client := httptest.NewServer(proxyHandler), http.Client{}
		proxyURL, _ := url.Parse(proxy.URL)

		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL), TLSClientConfig: clientTLS}

		defer proxy.Close()

		stub := newStubSvrFunc(http.HandlerFunc(func(rs http.ResponseWriter, rcvRq *http.Request) {
			b, _ := io.ReadAll(rcvRq.Body)
			rcvMethod, rcvPath, rcvQSV, rcvHdrV, rcvIntHdrV, rcvBdy = rcvRq.Method, rcvRq.URL.Path, rcvRq.URL.Query().Get(rqQSK), rcvRq.Header.Get(rqHdrK), rcvRq.Header.Get(intRqHdrK), string(b)

			rs.Header().Add(rsHdrK, rsHdrV)
			rs.WriteHeader(rsCode)
			rs.Write([]byte(rsBdy))
		}))

		defer stub.Close()

		rq, _ := http.NewRequest(rqMethod, fmt.Sprintf("%v/%v/?%v=%v", stub.URL, rqPath, rqQSK, rqQSV), strings.NewReader(rqBdy))
		rq.Header.Set(rqHdrK, rqHdrV)

		if icpt {
			id := SetIntercept(intercept.NewIntercept("test-intercept",
				intercept.MatchAllRequests,
				intercept.MatchAllResponses,
				func(r *intercept.ProxyRequest) error {
					r.Header.Add(intRqHdrK, intRqHdrV)
					return nil
				},
				func(r *intercept.ProxyResponse) error {
					r.Header.Add(intRsHdrK, intRsHdrV)
					return nil
				},
			))

			defer UnsetIntercept(id)
		}

		rs, err := client.Do(rq)

		if err != nil {
			t.Fatalf("expected no error proxying request, got [%v]", err)
		}

		assert := func(attr, exp, got string) {
			if got != exp {
				t.Fatalf("expected to %v [%v], got [%v]", attr, exp, got)
			}
		}

		assert("receive request method of", rqMethod, rcvMethod)
		assert("receive request path of", "/"+rqPath+"/", rcvPath)
		assert("receive request querystring value of", rqQSV, rcvQSV)
		assert("receive request header value of", rqHdrV, rcvHdrV)
		assert("receive request body of", rqBdy, rcvBdy)
		assert("receive response header value of", rsHdrV, rs.Header.Get(rsHdrK))

		b, err := io.ReadAll(rs.Body)
		assert("receive response body of", rsBdy, string(b))

		if icpt {
			assert("receive intercepted request header value of", intRqHdrV, rcvIntHdrV)
			assert("receive intercepted response header value of", intRsHdrV, rs.Header.Get(intRsHdrK))
		}
	}

	tlsCfg := &tls.Config{InsecureSkipVerify: true}

	t.Run("HTTP", func(t *testing.T) { test(t, false, nil, HTTPHandler(), httptest.NewServer) })
	t.Run("InterceptedHTTP", func(t *testing.T) { test(t, true, nil, HTTPHandler(), httptest.NewServer) })
	t.Run("HTTPS", func(t *testing.T) { test(t, false, tlsCfg, HTTPSHandler(), httptest.NewTLSServer) })
	t.Run("InterceptedHTTPS", func(t *testing.T) { test(t, true, tlsCfg, HTTPSHandler(), httptest.NewTLSServer) })
}
