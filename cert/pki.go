package cert

import (
	"bytes"
	"comradequinn/hflow/log"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

// ca extracts the X509 certificate and private key of the HFLOW CA from local PEM files
func ca() (*x509.Certificate, *rsa.PrivateKey) {
	readPEM := func(bytes []byte, pemType string) []byte {
		pemBlock, _ := pem.Decode(bytes)

		if pemBlock == nil {
			log.Fatalf(0, "expected pem [%v] data but data is not in pem format", pemType)
		}

		if pemBlock.Type != pemType {
			log.Fatalf(0, "expected pem data for type [%v] but found [%v]", pemType, pemBlock.Type)
		}

		return pemBlock.Bytes
	}

	var err error
	var privateKey *rsa.PrivateKey

	if privateKey, err = x509.ParsePKCS1PrivateKey(readPEM(hflowCAPrivateKeyPEM, "RSA PRIVATE KEY")); err != nil {
		log.Fatalf(0, "unable to parse hflow ca private key pem as pkcs1 private key [%v]. ", err)
	}

	var pkixPublicKey interface{}

	if pkixPublicKey, err = x509.ParsePKIXPublicKey(readPEM(hflowCAPublicKeyPEM, "PUBLIC KEY")); err != nil {
		log.Fatalf(0, "unable to parse hflow ca public key pem as pkix public key [%v]", err)
	}

	var rsaPublicKey *rsa.PublicKey
	var ok bool

	if rsaPublicKey, ok = pkixPublicKey.(*rsa.PublicKey); !ok {
		log.Fatalf(0, "unable to parse hflow ca public key pkix data as rsa public key [%v]", err)
	}

	privateKey.PublicKey = *rsaPublicKey

	log.Printf(3, "loaded hflow ca key public and private keys")

	var certificate *x509.Certificate

	if certificate, err = x509.ParseCertificate(readPEM(hflowCACertPEM, "CERTIFICATE")); err != nil {
		log.Fatalf(0, "unable to parse hflow ca cert pem data as x509 certificate [%v]", err)
	}

	log.Printf(3, "loaded hflow ca certificate")

	return certificate, privateKey
}

func newEECert(subject string, ip net.IP, caCertificate *x509.Certificate, caPrivateKey *rsa.PrivateKey) (*tls.Certificate, error) {
	log.Printf(3, "generating end entity certificate for subject [%v]", subject)

	now, serialNumberLimit := time.Now(), new(big.Int).Lsh(big.NewInt(1), 128)

	template := x509.Certificate{
		Subject: pkix.Name{
			Organization:  []string{"HFLOW Dynamic Cert"},
			StreetAddress: []string{"1 Virtual Avenue"},
			PostalCode:    []string{"12345"},
			Province:      []string{"Ether"},
			Locality:      []string{"Net"},
			Country:       []string{"UK"},
			CommonName:    subject,
		},
		NotBefore:             now,
		NotAfter:              now.Add(87658 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		IsCA:                  false,
		BasicConstraintsValid: true,
		DNSNames:              []string{subject},
		IPAddresses:           []net.IP{ip},
	}

	var err error

	if template.SerialNumber, err = rand.Int(rand.Reader, serialNumberLimit); err != nil {
		return nil, fmt.Errorf("failed to generate serial number for end entity certificate with subject [%v]. [%v]", subject, err)
	}

	log.Printf(3, "generated serial number of [%v] for end entity certificate with subject [%v]", template.SerialNumber, subject)

	var pk *rsa.PrivateKey

	if pk, err = rsa.GenerateKey(rand.Reader, 4096); err != nil {
		return nil, fmt.Errorf("failed to generate private key for end entity certificate with subject [%v]. [%v]", subject, err)
	}

	log.Printf(3, "generated private key for end entity certificate with subject [%v]", subject)

	var der []byte

	if der, err = x509.CreateCertificate(rand.Reader, &template, caCertificate, &pk.PublicKey, caPrivateKey); err != nil {
		return nil, fmt.Errorf("failed to generate der encoded end entity certificate with subject [%v]. [%v]", subject, err)
	}

	certPEM, keyPEM := bytes.Buffer{}, bytes.Buffer{}

	if err = pem.Encode(&certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: der}); err != nil {
		return nil, fmt.Errorf("failed to transcode end entity certificate with subject [%v] from der to pem [%v]", subject, err)
	}

	if err = pem.Encode(&keyPEM, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pk)}); err != nil {
		return nil, fmt.Errorf("failed to encode private key for end entity certificate with subject [%v] to pem [%v]", subject, err)
	}

	var cert tls.Certificate

	if cert, err = tls.X509KeyPair(certPEM.Bytes(), keyPEM.Bytes()); err != nil {
		return nil, fmt.Errorf("failed to generate end entity certificate with subject [%v] from pem [%v]", subject, err)
	}

	log.Printf(2, "generated end entity certificate with subject [%v]", subject)

	return &cert, nil
}
