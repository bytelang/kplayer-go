package types

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/forgoer/openssl"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
)

var (
	TlsRootCert    = ""
	TlsClientToken = ""
	CipherKey      = ""
	CipherIV       = ""
)
var (
	RawTlsRootCert   []byte
	RawTlsClientCert []byte
	RawTlsClientKey  []byte
)

func init() {
	RawTlsRootCert = TlsSecretDecode(TlsRootCert)
	if err := LoadClientToken(TlsClientToken); err != nil {
		log.Fatalf("cannot load client default token: %s", err)
	}
}

type TlsClient struct {
}

func GetTlsClientConfig() (*tls.Config, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	if ok := pool.AppendCertsFromPEM(RawTlsRootCert); !ok {
		return nil, fmt.Errorf("certification append error")
	}

	cert, err := tls.X509KeyPair(RawTlsClientCert, RawTlsClientKey)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		RootCAs:      pool,
		Certificates: []tls.Certificate{cert},
	}, nil
}

func LoadClientToken(tokenData string) error {
	cert, key, err := TlsClientDecode(tokenData)
	if err != nil {
		return err
	}

	RawTlsClientCert = cert
	RawTlsClientKey = key

	return nil
}

func TlsSecretDecode(data string) []byte {
	hexRawData, err := hex.DecodeString(data)
	if err != nil {
		log.Fatal(err)
	}
	rawData, err := openssl.AesCBCDecrypt(hexRawData, []byte(CipherKey), []byte(CipherIV), openssl.PKCS5_PADDING)
	if err != nil {
		log.Fatal(err)
	}

	return rawData
}

func TlsClientDecode(data string) ([]byte, []byte, error) {
	// hex decode
	rawHexData, err := hex.DecodeString(data)
	if err != nil {
		return nil, nil, err
	}

	// gzip unpack
	gr, err := gzip.NewReader(bytes.NewReader(rawHexData))
	if err != nil {
		return nil, nil, err
	}
	defer gr.Close()

	rawData, err := ioutil.ReadAll(gr)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, nil, err
	}

	dataStruct := make(map[string][]byte)
	if err := json.Unmarshal(rawData, &dataStruct); err != nil {
		return nil, nil, err
	}

	cert, ok := dataStruct["cert"]
	if !ok {
		return nil, nil, fmt.Errorf("tls client data invalid")
	}
	key, ok := dataStruct["key"]
	if !ok {
		return nil, nil, fmt.Errorf("tls client data invalid")
	}

	return cert, key, nil
}
