package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

type request struct {
	//todo
}

var debug bool

func handleRequest(req interface{} /*request*/) (string, error) {
	dirURL := os.Getenv("DIRECTORY_URL")
	if dirURL == "" {
		return "", fmt.Errorf("DIRECTORY_URL env var missing")
	}
	cli := NewCertClient(dirURL, os.Getenv("TOS_URL")) //todo okay to use "" if unset?

	certFile, isSet := os.LookupEnv("SERVER_CERT")
	if isSet {
		pool, err := x509.SystemCertPool()
		if err != nil {
			return "", fmt.Errorf("SystemCertPool: %w", err)
		}

		pemBytes, err := ioutil.ReadFile(certFile)
		if err != nil {
			return "", fmt.Errorf("Read cert file: %w", err)
		}

		ok := pool.AppendCertsFromPEM(pemBytes)
		if !ok {
			return "", fmt.Errorf("AppendCertsFromPEM returned false")
		}

		cli.HTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: pool,
				},
			},
		}
	}

	err := cli.GetCert("test01.bitflip.space")
	return "", err
}

func init() {
	debug = os.Getenv("DEBUG") != ""
}

func main() {
	if debug {
		s, e := handleRequest(request{})
		fmt.Printf("result: %v\nerror: %v\n", s, e)
	} else {
		lambda.Start(handleRequest)
	}
}
