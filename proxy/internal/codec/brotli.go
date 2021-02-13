package codec

import (
	"bytes"
	"fmt"
	"io/ioutil"

	br "github.com/andybalholm/brotli"
)

type brotli struct{}

func (brotli) dec(data []byte) ([]byte, error) {
	r := br.NewReader(bytes.NewReader(data))

	b, err := ioutil.ReadAll(r)

	if err != nil {
		return nil, fmt.Errorf("unable to decompress data with brotli: [%v]", err)
	}

	return b, nil
}

func (brotli) enc(data []byte) ([]byte, error) {
	b := bytes.Buffer{}
	w := br.NewWriter(&b)

	_, err := w.Write(data)

	if err != nil {
		return nil, fmt.Errorf("unable to compress data with brotli: [%v]", err)
	}

	if err = w.Close(); err != nil {
		return nil, fmt.Errorf("unable to close brotli writer: [%v]", err)
	}

	return b.Bytes(), nil
}
