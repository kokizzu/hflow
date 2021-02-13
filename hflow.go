package main

import (
	"comradequinn/hflow/cert"
	"comradequinn/hflow/log"
	"comradequinn/hflow/proxy"
	"comradequinn/hflow/proxy/intercept"
	"comradequinn/hflow/syncio"
	"flag"
	"fmt"
	"net/http"
	"os"
)

func main() {
	caExport, proxyHTTPPort, proxyHTTPSPort := false, 0, 0

	flag.BoolVar(&caExport, "ca", false, "write the hflow ca certificate in pem format to stdout and exit")
	flag.IntVar(&proxyHTTPPort, "p", 8080, "the port to proxy http over")
	flag.IntVar(&proxyHTTPSPort, "ps", 4443, "the port to proxy https over")

	url := flag.String("u", "", "only capture requests that contain the url-pattern. ignored if --api is set")
	status := flag.String("s", "", "only capture responses that contain the status-pattern. ignored if --api not set")
	binary := flag.Bool("b", false, "write non-text response bodies")
	limit := flag.Int("l", -1, "limit text response bodies to the specified byte count when sending to writers, -1 is no limit")
	verbosity := flag.Int("v", 0, "the verbosity of the log output")

	flag.Parse()

	log.SetVerbosity(*verbosity)

	if caExport {
		if err := cert.WriteCA(os.Stdout); err != nil {
			log.Printf(0, "error writing hflow ca certificate [%v]", err)
		}

		log.Printf(0, "hflow ca certificate written to stdout in pem format")

		return
	}

	log.Printf(0, "response body limit set at [%v] bytes", *limit)

	mrq := intercept.MatchRequestURL(*url)

	proxy.SetIntercept(intercept.Writer("stdout writer", mrq, intercept.MatchResponseStatus(*status, mrq), *binary, *limit, syncio.NewWriter(os.Stdout)))

	startSvr := func(name string, port int, handler http.Handler) {
		svr := http.Server{
			Addr:    fmt.Sprintf(":%v", port),
			Handler: handler,
		}

		go func() {
			if err := svr.ListenAndServe(); err != nil {
				log.Fatalf(0, "error starting proxy server on port [%v]: [%v]", port, err)
			}
		}()

		log.Printf(0, "%v started on port [%v]", name, port)

	}

	startSvr("http proxy server", proxyHTTPPort, proxy.HTTPHandler())
	startSvr("https proxy server", proxyHTTPSPort, proxy.HTTPSHandler())

	select {}
}
