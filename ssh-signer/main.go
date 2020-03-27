package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

type request struct {
	Arn       string
	PublicKey string
}

var outBucket string
var outPathPrefix string

var debugMode bool
var initError error
var signer *Signer
var arnPattern *regexp.Regexp
var awsCli *awsClient

func handleRequest(req request) (string, error) {
	if initError != nil {
		// A panic in the handler causes the Lambda framework to return an error and kill this instance.
		fmt.Fprintf(os.Stderr, "init error: %s\n", initError)
		panic(initError)
	}

	if req.Arn == "" {
		return "", fmt.Errorf("missing Arn in request")
	}
	if req.PublicKey == "" {
		return "", fmt.Errorf("missing PublicKey in request")
	}

	matches := arnPattern.FindStringSubmatch(req.Arn)
	if matches == nil {
		return "", fmt.Errorf("failed to parse Arn in request")
	}

	role := matches[1]
	instance := matches[2]
	roleID, err := awsCli.getRoleID(role)
	if err != nil {
		return "", fmt.Errorf("get role id: %w", err)
	}

	outCert, err := signer.MakeCertificate(req.PublicKey, []string{instance, role})
	if err != nil {
		return "", fmt.Errorf("make certificate: %w", err)
	}

	outPath := fmt.Sprintf("%s/%s:%s/certificate", outPathPrefix, roleID, instance)

	err = awsCli.putS3Object(outBucket, outPath, outCert)
	if err != nil {
		// log and discard error details to prevent any data leak to client
		fmt.Printf("error putting cert in S3: %s", err)
		return "", fmt.Errorf("failed to put cert")
	}

	outURI := fmt.Sprintf("s3://%s/%s", outBucket, outPath)

	return outURI, nil
}

func parseS3URI(uri string) (string, string, error) {
	if uri == "" {
		return "", "", fmt.Errorf("missing / empty")
	}

	parts := strings.SplitN(uri, "/", 4)
	if len(parts) < 4 {
		return "", "", fmt.Errorf("too few components")
	}

	if parts[0] != "s3:" || parts[1] != "" {
		return "", "", fmt.Errorf("expected s3:// prefix")
	}

	return parts[2], parts[3], nil
}

// We defer any initialization errors until the handler is called so we can return an error
// message to the caller instead of just timing out.
func init() {
	debugMode = (os.Getenv("DEBUG") != "")

	arnPattern = regexp.MustCompile("\\Aarn:aws:sts::\\d+:assumed-role/([a-z0-9_-]+)/(i-[0-9a-f]+)\\z")

	caKeyBucket, caKeyPath, err := parseS3URI(os.Getenv("CA_KEY_URI"))
	if err != nil {
		initError = fmt.Errorf("CA_KEY_URI: %w", err)
		return
	}

	ob, op, err := parseS3URI(os.Getenv("OUTPUT_URI_PREFIX"))
	if err != nil {
		initError = fmt.Errorf("OUTPUT_URI_PREFIX: %w", err)
		return
	}
	outBucket = ob
	outPathPrefix = op

	aws, err := newAwsClient()
	if err != nil {
		initError = fmt.Errorf("initializing aws client: %w", err)
		return
	}
	awsCli = aws

	caKey, err := awsCli.getS3Object(caKeyBucket, caKeyPath)
	if err != nil {
		fmt.Printf("error getting CA key, bucket=%q path=%q\n", caKeyBucket, caKeyPath)
		initError = fmt.Errorf("getting caKey: %w", err)
		return
	}

	signer, err = NewSigner(caKey)
	if err != nil {
		initError = fmt.Errorf("creating signer: %w", err)
		return
	}

	if debugMode {
		fmt.Printf("init completed")
	}
}

func main() {
	if debugMode {
		fmt.Printf("initError=%v\n", initError)
		fmt.Printf("signer=%v\n", signer)

		testKeyBytes, err := ioutil.ReadFile(os.Args[2])
		if err != nil {
			panic(err)
		}
		s, e := handleRequest(request{
			Arn:       os.Args[1],
			PublicKey: string(testKeyBytes),
		})
		fmt.Printf("result: %v\nerror: %v\n", s, e)
	} else {
		fmt.Println("Starting Lambda framework")
		lambda.Start(handleRequest)
	}
}
