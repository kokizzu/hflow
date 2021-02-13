package intercept

import "strings"

// MatchRequestFunc describes a func which should return (true, nil)
// when passed a *Request in order for an action to be applied
type MatchRequestFunc func(*ProxyRequest) (bool, error)

// MatchResponseFunc describes a func which should return (true, nil) when passed
// the *Request & *Response in order for an action to be applied
type MatchResponseFunc func(*ProxyRequest, *ProxyResponse) (bool, error)

// RequestFunc describes a func that performs an action on the specified request.
type RequestFunc func(*ProxyRequest) error

// ResponseFunc describes a func that performs an action on the specified response.
type ResponseFunc func(*ProxyResponse) error

var (
	// MatchAllRequests matches any request
	MatchAllRequests MatchRequestFunc = func(r *ProxyRequest) (bool, error) { return true, nil }

	// MatchAllResponses matches any response
	MatchAllResponses MatchResponseFunc = func(r *ProxyRequest, rs *ProxyResponse) (bool, error) { return true, nil }

	// MatchRequestURL returns a MatchRequestFunc that matches on requests with a URL that contains s
	MatchRequestURL = func(s string) MatchRequestFunc {
		if s == "" {
			return MatchAllRequests
		}

		return func(r *ProxyRequest) (bool, error) {
			return strings.Contains(r.URL.String(), s), nil
		}
	}

	// MatchResponseStatus returns a MatchResponseFunc that matches responses with a status
	// that contains s and a source request that matches MatchRequestFunc
	MatchResponseStatus = func(s string, mrq MatchRequestFunc) MatchResponseFunc {
		rsMatchFunc := func(string, *ProxyResponse) bool { return true }

		if s != "" {
			rsMatchFunc = func(s string, rs *ProxyResponse) bool { return strings.Contains(rs.Status, s) }
		}

		return func(r *ProxyRequest, rs *ProxyResponse) (bool, error) {
			rqMatch, err := mrq(r)

			if !rqMatch || err != nil {
				return rqMatch, err
			}

			return rsMatchFunc(s, rs), nil
		}
	}
)
