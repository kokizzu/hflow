// Package intercept provides functions for creating and configuring intercepts with package proxy
package intercept

import (
	"comradequinn/hflow/log"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Writer writes
// * request traffic to the specified io.Writer where the mrq matches the request
// * response traffic to the specified io.Writer where mrs matches the response
//
// Unless binary is set to true, only text-based mime-type bodies are written to
// If limit is greater than or equal to 0, then text response body writes are capped at that number of bytes
func Writer(label string, mrq MatchRequestFunc, mrs MatchResponseFunc, binary bool, limit int, w io.Writer) *Intercept {
	writeHTTP := func(h http.Header, b []byte, sb *strings.Builder) error {
		const delim = "__________________________________________________________________________________________________________\n\n"
		contentType, textContentTypes := "", []string{"text/", "/json", "xml", "/javascript", "urlencoded"}

		for k, vs := range h {
			if k == "Content-Type" {
				contentType = strings.Split(vs[0], ";")[0]
			}

			sb.WriteString(fmt.Sprintf("%v: ", k))
			sb.WriteString(strings.Join(vs, ","))

			sb.WriteString("\n")
		}

		if limit >= 0 && len(b) > limit {
			b = b[:limit]
		}

		body := string(b)

		if contentType != "" {
			text := false

			for _, tct := range textContentTypes {
				if strings.Contains(contentType, tct) {
					text = true
					break
				}
			}

			if !text && !binary {
				body = "[binary data]"
			}
		}

		go func() {
			if len(body) > 0 {
				sb.WriteString("\n" + body + "\n")
			}

			sb.WriteString(delim)

			if _, err := w.Write([]byte(sb.String())); err != nil {
				log.Printf(0, "unable to write to io.Writer during writer intercept labelled [%v]: [%v]", label, err)
			}
		}()

		return nil
	}

	return NewIntercept(label, mrq, mrs,
		func(r *ProxyRequest) error {
			sb := strings.Builder{}

			sb.WriteString(fmt.Sprintf(">>> %v %v\n\n", r.Method, r.URL.String()))

			err := writeHTTP(r.Header, r.Body, &sb)

			if err != nil {
				return fmt.Errorf("unable to read request for [%v] in writer intercept labelled [%v]: [%v]", r.URL.String(), label, err)
			}

			return nil
		},
		func(r *ProxyResponse) error {
			sb := strings.Builder{}

			sb.WriteString(fmt.Sprintf("<<< %v from %v %v\n\n", r.Status, r.Request.Method, r.Request.URL.String()))

			err := writeHTTP(r.Header, r.Body, &sb)

			if err != nil {
				return fmt.Errorf("unable to read response to [%v] in writer intercept labelled [%v]: [%v]", r.Request.URL.String(), label, err)
			}

			return nil
		},
	)
}
