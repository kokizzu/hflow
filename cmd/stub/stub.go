package main

import (
	"comradequinn/hflow/cmd/stub/echo"
	"comradequinn/hflow/log"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
)

// Start begins the stub listening on the specified port
func main() {
	log.Printf(0, "starting")

	httpPort, httpsPort := 0, 0

	flag.IntVar(&httpPort, "port", 8081, "the port for the stub server to listen for http on")
	flag.IntVar(&httpsPort, "tls", 4444, "the port for the stub server to listen for https on")

	v := flag.Int("v", 0, "the verbosity of the log output")

	flag.Parse()

	log.SetVerbosity(*v)

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%v", httpPort), echo.StubHandler); err != nil {
			log.Fatalf(0, "error starting stub http server [%v]", err)
		}
	}()

	go func() {
		svr := &http.Server{
			Addr:    fmt.Sprintf(":%v", httpsPort),
			Handler: echo.StubHandler,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{echo.StubCert},
			},
		}

		if err := svr.ListenAndServeTLS("", ""); err != nil {
			log.Fatalf(0, "error starting stub https server [%v]", err)
		}
	}()

	log.Printf(0, "stub server started on port [%v] for http and [%v] for https", httpPort, httpsPort)

	select {}
}
