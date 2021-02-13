package proxy

import (
	"comradequinn/hflow/log"
	"comradequinn/hflow/proxy/intercept"
	"sync"
)

var lockIntercepts = func() func(f func(map[int]*intercept.Intercept, *int), readonly bool) {
	intercepts := map[int]*intercept.Intercept{}
	id := 0
	mx := sync.RWMutex{}

	return func(f func(map[int]*intercept.Intercept, *int), readonly bool) {
		if readonly {
			mx.RLock()
			defer mx.RUnlock()
		} else {
			mx.Lock()
			defer mx.Unlock()
		}

		f(intercepts, &id)
	}
}()

// SetIntercept creates and applies a new http traffic interception based on the specified arguments. Label has no
// no programmatic purpose, serving only to describe the interception to clients and as such
// can be any value
func SetIntercept(i *intercept.Intercept) int {
	iid := 0

	if i == nil {
		log.Panicf(0, "cannot set nil as an intercept")
	}

	lockIntercepts(func(intercepts map[int]*intercept.Intercept, id *int) {
		*id++
		iid = *id
		intercepts[*id] = i
	}, false)

	log.Printf(1, "added intercept labelled [%v]", i.Label())

	return iid
}

// UnsetIntercept causes the specified intercept to cease being applied to http traffic
func UnsetIntercept(id int) {
	lockIntercepts(func(intercepts map[int]*intercept.Intercept, _ *int) { delete(intercepts, id) }, false)

	log.Printf(1, "removed intercept labelled [%v]", id)
}

// Intercepts returns all configured intercepts
func Intercepts() map[int]*intercept.Intercept {
	copy := map[int]*intercept.Intercept{}

	lockIntercepts(func(intercepts map[int]*intercept.Intercept, _ *int) {
		for id, intercept := range intercepts {
			i := *intercept
			copy[id] = &i
		}
	}, true)

	return copy
}
