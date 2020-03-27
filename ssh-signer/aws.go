package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
	"io/ioutil"
	"strings"
)

type awsClient struct {
	session   *session.Session
	s3Client  *s3.S3
	iamClient *iam.IAM
}

func newAwsClient() (*awsClient, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("new session: %w", err)
	}

	return &awsClient{
		session:   sess,
		s3Client:  s3.New(sess),
		iamClient: iam.New(sess),
	}, nil
}

func (a *awsClient) getS3Object(bucket, path string) (string, error) {
	if a == nil {
		return "", fmt.Errorf("getS3Object called on nil aws instance")
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	}

	res, err := a.s3Client.GetObject(input)
	if err != nil {
		return "", fmt.Errorf("get object: %w", err)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read object body: %w", err)
	}

	return string(bytes), nil
}

func (a *awsClient) putS3Object(bucket, path, content string) error {
	if a == nil {
		return fmt.Errorf("putS3Object called on nil aws instance")
	}

	input := &s3.PutObjectInput{
		Body:   strings.NewReader(content), // aws.ReadSeekCloser(
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	}

	_, err := a.s3Client.PutObject(input)
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}

	return nil
}

func (a *awsClient) getRoleID(name string) (string, error) {
	if a == nil {
		return "", fmt.Errorf("getRoleID called on nil aws instance")
	}

	input := &iam.GetRoleInput{
		RoleName: aws.String(name),
	}

	res, err := a.iamClient.GetRole(input)
	if err != nil {
		return "", fmt.Errorf("get role: %w", err)
	}

	return *res.Role.RoleId, nil
}
