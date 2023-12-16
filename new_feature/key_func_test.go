package new_feature

import (
	"encoding/base64"
	"testing"
)

func TestEncode(t *testing.T) {
	pubKey := "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE/17XRshxRD8s6SL4LmvMT0Dvu6NhKqRP1FUPMR6n1xCN8v9An/8gJp0Oe6GOxgzbCtj/QmU6ZKfGgxMBbDe9vg=="
	encoded, err := Encode(pubKey, []byte("MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE/17XRshxRD8s6SL4LmvMT0Dvu6NhKqRP1FUPMR6n1xCN8v9An/8gJp0Oe6GOxgzbCtj/QmU6ZKfGgxMBbDe9vg==MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE/17XRshxRD8s6SL4LmvMT0Dvu6NhKqRP1FUPMR6n1xCN8v9An/8gJp0Oe6GOxgzbCtj/QmU6ZKfGgxMBbDe9vg==MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE/17XRshxRD8s6SL4LmvMT0Dvu6NhKqRP1FUPMR6n1xCN8v9An/8gJp0Oe6GOxgzbCtj/QmU6ZKfGgxMBbDe9vg==MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE/17XRshxRD8s6SL4LmvMT0Dvu6NhKqRP1FUPMR6n1xCN8v9An/8gJp0Oe6GOxgzbCtj/QmU6ZKfGgxMBbDe9vg=="))
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("encode res:", encoded)
	}
	priKey := "e+69ZcM25oSOwDH5nQk9kyM7myLu7rxuJvGrOQECOl0="
	//priBytes, _ := base64.StdEncoding.DecodeString(priKey)
	decoded, err := Decode(priKey, encoded)
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log(string(decoded))
	}
}
func TestEncode1(t *testing.T) {
	encoded, err := base64.StdEncoding.DecodeString(`BJXNxUOpfr7vQ+jWsmF8UAd9ufvNAkm+wLbrjj/D8udhncXqFZkx8i/UBc2KWJV1TzpQmqCqlhba
r+nWKaJFqCSdM/8mzxSkUdB4tJeOmX0g7WjSvS+fnsoVzxYLVhO1ZCH8sGD3Dk6QpwtP3y0rya+R
EA==
`)
	priKey := "0SEeDatf+3jkR3pR7MldjpnfYWbPsKzKMdWLRt7xG+A="
	//priBytes, _ := base64.StdEncoding.DecodeString(priKey)
	decoded, err := Decode(priKey, encoded)
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log(string(decoded))
	}
}
