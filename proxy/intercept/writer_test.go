package intercept

import (
	"net/http"
	"strings"
	"testing"
)

func TestWriter(t *testing.T) {
	test := func(rsContentType string, expectBinaryResponse bool) {
		rqHdrK, rqHdrV, rqbody, rsHdrK, rsHdrV, rsbody, tb := "Rq-Hk", "Rq-Hv", "rq-body-data", "Rs-Hk", "Rs-Hv", "rs-body-data", &TestBuffer{Wrote: make(chan struct{}, 2)}

		rq, _ := http.NewRequest(http.MethodPost, "http://www.test.com/echo/?data=some-qs-data", strings.NewReader(rqbody))
		rq.Header.Set(rqHdrK, rqHdrV)

		prq, _ := newProxyRequest(rq)

		i, prs := Writer("testwriter",
			MatchAllRequests,
			MatchAllResponses,
			false, -1, tb), &ProxyResponse{Status: "200 OK", Header: http.Header{rsHdrK: []string{rsHdrV}, "Content-Type": []string{rsContentType}}, Body: []byte(rsbody), Request: rq}

		if err := i.request(prq); err != nil {
			t.Fatalf("expected no error processing request, got [%v]", err)
		}

		if err := i.response(prs); err != nil {
			t.Fatalf("expected no error processing response, got [%v]", err)
		}

		<-tb.Wrote
		<-tb.Wrote

		if !strings.Contains(tb.Buffer.String(), rq.URL.String()) {
			t.Fatalf("expected output to contain url [%v], got [%v]", rq.URL.String(), tb.Buffer.String())
		}

		if !strings.Contains(tb.Buffer.String(), rqHdrK) || !strings.Contains(tb.Buffer.String(), rqHdrV) {
			t.Fatalf("expected output to contain request header [%v:%v], got [%v]", rqHdrK, rqHdrV, tb.Buffer.String())
		}

		if !strings.Contains(tb.Buffer.String(), rq.Method) {
			t.Fatalf("expected output to contain method [%v], got [%v]", rq.Method, tb.Buffer.String())
		}

		if !strings.Contains(tb.Buffer.String(), rqbody) {
			t.Fatalf("expected output to contain request body [%v], got [%v]", rqbody, tb.Buffer.String())
		}

		if !strings.Contains(tb.Buffer.String(), prs.Status) {
			t.Fatalf("expcted output to contain response status [%v], got [%v]", rqbody, tb.Buffer.String())
		}

		if !strings.Contains(tb.Buffer.String(), rsHdrK) || !strings.Contains(tb.Buffer.String(), rsHdrV) {
			t.Fatalf("expected output to contain response header [%v:%v], got [%v]", rsHdrK, rsHdrV, tb.Buffer.String())
		}

		if expectBinaryResponse && !strings.Contains(tb.Buffer.String(), "[binary data]") {
			t.Fatalf("expected output to contain response body [%q], got [%v]", "[binary data]", tb.Buffer.String())
		}

		if !expectBinaryResponse && !strings.Contains(tb.Buffer.String(), rsbody) {
			t.Fatalf("expected output to contain response body [%v], got [%v]", rsbody, tb.Buffer.String())
		}
	}

	test("text/plain", false)
	test("image/gif", true)
}
