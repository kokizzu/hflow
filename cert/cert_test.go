package cert

import (
	"crypto/tls"
	"strconv"
	"testing"
)

// noOptimize is used to assign unneeded results to so as to
// prevent the compiler attempting to optimising the program
// by removing a seemingly redundant function call
var noOptimize interface{}

func Benchmark_Get(b *testing.B) {
	noOptimize, _ = Get(&tls.ClientHelloInfo{ServerName: "warm-up.func"})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		noOptimize, _ = Get(&tls.ClientHelloInfo{ServerName: "domain" + strconv.Itoa(b.N%10) + ".com"})
	}
}

func TestCertGet(t *testing.T) {
	_, err := Get(&tls.ClientHelloInfo{ServerName: "domain.com"})

	if err != nil {
		t.Fatalf("expected no error after cert generation, got: [%v]", err)
	}
}
