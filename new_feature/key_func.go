package new_feature

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/matrix-org/gomatrixserverlib"
	"strings"
)

var priBase64 string
var servername string
var federationClient *gomatrixserverlib.FederationClient

func Init(pri string, cfgServername string, federation *gomatrixserverlib.FederationClient) {
	priBase64 = pri
	servername = cfgServername
	federationClient = federation
}

func GetPubEcKeyFromPri(priKey []byte) (string, error) {
	privateSerHex := hex.EncodeToString(priKey)
	//ecdsa
	prk, err := ecdsa.GenerateKey(elliptic.P256(), strings.NewReader(privateSerHex))
	ecdsaPubKey := ecies.ImportECDSA(prk).PublicKey
	ecdsaPubKeyBytes, err := x509.MarshalPKIXPublicKey(ecdsaPubKey.ExportECDSA())
	if err != nil {
		return "", errors.New("x509.MarshalPKIXPublicKey() failed:" + err.Error())
	}

	ecdsaPubKeyStr := base64.StdEncoding.EncodeToString(ecdsaPubKeyBytes)
	return ecdsaPubKeyStr, nil
}

// Encode eciesPublicKey
func Encode(eciesPublicKey string, msg []byte) ([]byte, error) {
	pukByte, err := base64.StdEncoding.DecodeString(eciesPublicKey)
	if err != nil {
		return []byte{}, err
	}
	pukInterface, err := x509.ParsePKIXPublicKey(pukByte)
	ecdsaPubKey := pukInterface.(*ecdsa.PublicKey)
	eciesPubKey := ecies.ImportECDSAPublic(ecdsaPubKey)

	return ECCEncrypt(msg, *eciesPubKey)
}
func ECCEncrypt(pt []byte, puk ecies.PublicKey) ([]byte, error) {
	ct, err := ecies.Encrypt(rand.Reader, &puk, pt, nil, nil)
	return ct, err
}

// Decode addr
func Decode(privateKey string, msg []byte) ([]byte, error) {
	bytesPri, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return nil, err
	}
	privateSerHex := hex.EncodeToString(bytesPri)
	prk, err := ecdsa.GenerateKey(elliptic.P256(), strings.NewReader(privateSerHex))
	if err != nil {
		return []byte{}, err
	}
	prk2 := ecies.ImportECDSA(prk)
	return ECCDecrypt(msg, *prk2)
}

// ServerDecode addr
func ServerDecode(msg string) ([]string, error) {
	if msg == "" {
		return []string{}, nil
	}
	msgBytes, err2 := base64.StdEncoding.DecodeString(msg)
	if err2 != nil {
		return nil, err2
	}
	bytesPri, err := base64.StdEncoding.DecodeString(priBase64)
	if err != nil {
		return nil, err
	}
	privateSerHex := hex.EncodeToString(bytesPri)
	prk, err := ecdsa.GenerateKey(elliptic.P256(), strings.NewReader(privateSerHex))
	if err != nil {
		return nil, err
	}
	prk2 := ecies.ImportECDSA(prk)
	eccDecrypt, err2 := ECCDecrypt(msgBytes, *prk2)
	if err2 != nil {
		return nil, err2
	} else {
		res := []string{}
		err = json.Unmarshal(eccDecrypt, &res)
		if err != nil {
			return nil, err
		} else {
			return res, nil
		}
	}
}

func ECCDecrypt(ct []byte, prk ecies.PrivateKey) ([]byte, error) {
	pt, err := prk.Decrypt(ct, nil, nil)
	return pt, err
}
