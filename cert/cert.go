package cert

import (
	"comradequinn/hflow/log"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
)

var (
	// Get returns an end entity certificate, signed by the HFLOW CA, that matches the domain or ip that was requested
	Get = func() func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
		log.Printf(1, "initialising dynamic certificate generation")

		conn, err := net.Dial("udp", "192.0.0.0:9999") // Doesn't matter that nothing is listening, we just need to get the IP it selects to connect with

		if err != nil {
			log.Fatalf(0, "unable to ascertain host ip [%v]", err)
		}

		defer conn.Close()

		type cacheEntry struct {
			cert *tls.Certificate
			mx   sync.Mutex
		}

		cache, cacheMx, hostIP := map[string]*cacheEntry{}, sync.RWMutex{}, conn.LocalAddr().(*net.UDPAddr).IP
		caCertificate, caPrivateKey := ca()

		return func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
			subject := chi.ServerName

			if subject == "" {
				log.Printf(3, "using ip as subject of end entity certificate due to absence of server name in client hello")
				subject = hostIP.String()
			}

			cacheMx.RLock()

			ce, ok := cache[subject]

			cacheMx.RUnlock()

			if !ok {
				cacheMx.Lock()
				ce, ok = cache[subject]

				if !ok {
					ce = &cacheEntry{mx: sync.Mutex{}}
					cache[subject] = ce
				}
				cacheMx.Unlock()

				log.Printf(3, "added end entity certificate cache entry for [%v]", subject)
			}

			if ce.cert == nil {
				ce.mx.Lock()

				if ce.cert == nil {
					var err error

					if ce.cert, err = newEECert(subject, hostIP, caCertificate, caPrivateKey); err != nil {
						ce.cert = nil
						ce.mx.Unlock()
						return nil, fmt.Errorf("unable to create certificate for [%v]: [%v]", subject, err)
					}
				}

				ce.mx.Unlock()

				log.Printf(3, "assigned new end entity certificate to end entity certificate cache entry for [%v]", subject)
			}

			log.Printf(2, "added end entity certificate cache entry for [%v]", subject)

			return ce.cert, nil
		}
	}()
)

// WriteCA writes the HFLOW CA X509 certificate in PEM format to the specified io.Writer
func WriteCA(w io.Writer) error {
	_, err := w.Write(hflowCACertPEM)
	return err
}
