package echo

import (
	"comradequinn/hflow/log"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	// StubCert is the TLS certificate used to negotiate TLS handshakes
	StubCert tls.Certificate

	// StubHandler is an echo handler that responds to `/echo/?data=[value] BODY: [value]` by returning each [value] in its own response body
	StubHandler http.HandlerFunc = func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("X-From", "hflow-stub")

		if r.URL.Path != "/echo/" {
			rw.WriteHeader(http.StatusNotFound)
			log.Printf(0, "x<< rejected request for unsupported path [%v]", r.URL.String())
			return
		}

		scheme := "http"

		if r.TLS != nil {
			scheme = "https"
		}

		log.Printf(0, "<<< received [%v] request for [%v]", scheme, r.URL.String())
		fmt.Fprintf(rw, "qs-data: %v\n", r.URL.RawQuery)

		var (
			b   []byte
			err error
		)

		if b, err = ioutil.ReadAll(r.Body); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			log.Printf(0, "x<< unable to read body of request for [%v]: [%v]", r.URL.String(), err)
		}

		fmt.Fprintf(rw, "body-data: %v\n", string(b))
	}
)

func init() {
	var err error
	StubCert, err = tls.X509KeyPair(stubTLSCert, stubTLSKey)

	if err != nil {
		log.Fatalf(0, "error reading stub https server certificate or key [%v]", err)
	}
}
