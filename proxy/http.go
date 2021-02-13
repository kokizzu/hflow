package proxy

import (
	"comradequinn/hflow/log"
	"comradequinn/hflow/proxy/intercept"
	"comradequinn/hflow/proxy/internal/copy"
	"net/http"
)

// HTTPHandler is is a http.HandlerFunc that acts as HTTP Proxy
func HTTPHandler() http.HandlerFunc {
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
	}

	return func(rw http.ResponseWriter, r *http.Request) {
		log.Printf(1, "<<< received proxy request for [%v] on host [%v]", r.URL.String(), r.Host)

		var err error

		r, err = intercept.Request(r, Intercepts())

		if err != nil {
			rw.WriteHeader(http.StatusServiceUnavailable)
			log.Printf(0, "error intercepting request for [%v] on host [%v]: [%v]", r.URL.String(), r.Host, err)
			return
		}

		log.Printf(2, ">>> requesting [%v] from host [%v]", r.URL.String(), r.Host)

		rs, err := client.Do(r)

		if err != nil {
			rw.WriteHeader(http.StatusServiceUnavailable)
			log.Printf(0, "error proxying request for [%v] on host [%v]: [%v]", r.URL.String(), r.Host, err)
			return
		}

		log.Printf(2, "<<< received [%v] in response to [%v] on [%v]", rs.StatusCode, r.URL.String(), r.Host)

		rs, err = intercept.Response(r, rs, Intercepts())

		if err != nil {
			rw.WriteHeader(http.StatusServiceUnavailable)
			log.Printf(0, "error intercepting response to [%v] on host [%v]: [%v]", r.URL.String(), r.Host, err)
			return
		}

		copy.Header(rs.Header, rw.Header())

		b, err := copy.CloserToBytes(&rs.Body)

		if err != nil {
			rw.WriteHeader(http.StatusServiceUnavailable)
			log.Printf(0, "error reading response body from [%v]: [%v]", r.URL.String(), err)
			return
		}

		if _, err = rw.Write(b); err != nil {
			log.Printf(0, "error writing response body from [%v] on [%v] to hflow client: [%v]", r.URL.String(), r.Host, err)
			return
		}

		log.Printf(2, ">>> wrote proxy response for [%v]", r.URL.String())
	}
}
