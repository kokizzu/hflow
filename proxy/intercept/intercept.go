package intercept

import (
	"comradequinn/hflow/log"
	"fmt"
	"net/http"
)

// Intercept describes an action to be taken on a http exchange
// and the conditions to be met in order for that action to be applied
type Intercept struct {
	label    string
	matchRq  MatchRequestFunc
	request  RequestFunc
	matchRs  MatchResponseFunc
	response ResponseFunc
}

// NewIntercept returns a new Intercept based on the passed arguments
func NewIntercept(label string, mrq MatchRequestFunc, mrs MatchResponseFunc, rqf RequestFunc, rsf ResponseFunc) *Intercept {
	log.Printf(3, "creating intercept with label [%v]", label)

	setIfNil := func(f interface{}, df interface{}) {
		if f == nil {
			f = df
		}
	}

	setIfNil(mrq, func(_ *http.Request) (bool, error) { return false, nil })
	setIfNil(rqf, func(r *http.Request) (*http.Request, error) { return r, nil })
	setIfNil(mrs, func(_ *http.Request, _ *http.Response) (bool, error) { return false, nil })
	setIfNil(rsf, func(r *http.Response) (*http.Response, error) { return r, nil })

	return &Intercept{label: label, matchRq: mrq, request: rqf, matchRs: mrs, response: rsf}
}

// Label returns the Label of the Intercept
func (i *Intercept) Label() string {
	return i.label
}

// Request returns a new *http.Request which is the result of applying any matching intercepts to hr
func Request(hr *http.Request, intercepts map[int]*Intercept) (*http.Request, error) {
	log.Printf(3, "intercepting request for [%v]", hr.URL.String())

	r, err := newProxyRequest(hr)

	if err != nil {
		return nil, fmt.Errorf("error creating proxy request from https request to remote client [%v]. [%v]", hr.URL.String(), err)
	}

	matched := false

	for _, intercept := range intercepts {
		if matched, err = intercept.matchRq(r); err != nil {
			return nil, fmt.Errorf("error matching intercept [%v] to request for [%v]: [%v]", intercept.label, r.URL.String(), err)
		}

		if matched {
			log.Printf(2, "applying intercept labelled [%v] to request for [%v]", intercept.label, r.URL.String())

			if err = intercept.request(r); err != nil {
				return nil, fmt.Errorf("error applying intercept [%v] to request for [%v]: [%v]", intercept.label, r.URL.String(), err)
			}
		}
	}

	return r.http()
}

// Response returns a new *http.Response which is the result of applying any matching intercepts to hrs
func Response(hr *http.Request, hrs *http.Response, intercepts map[int]*Intercept) (*http.Response, error) {
	log.Printf(3, "interupting response for [%v]", hr.URL.String())

	rs, err := newProxyResponse(hrs)

	if err != nil {
		return nil, fmt.Errorf("error creating proxy response from https response to [%v]: [%v]", hr.URL.String(), err)
	}

	r, err := newProxyRequest(hr)

	if err != nil {
		return nil, fmt.Errorf("error creating proxy request from https request to remote client [%v]. [%v]", hr.URL.String(), err)
	}

	matched := false

	for _, intercept := range intercepts {
		if matched, err = intercept.matchRs(r, rs); err != nil {
			return nil, fmt.Errorf("error matching intercept [%v] to response to [%v]: [%v]", intercept.label, hr.URL.String(), err)
		}

		if matched {
			log.Printf(2, "applying intercept labelled [%v] to response to [%v]", intercept.label, hr.URL.String())

			if err = intercept.response(rs); err != nil {
				return nil, fmt.Errorf("error applying intercept [%v] to response to [%v]: [%v]", intercept.label, hr.URL.String(), err)
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("error creating proxy request from https request to remote client [%v]. [%v]", hr.URL.String(), err)
	}

	return rs.http()
}
