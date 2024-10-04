package storage

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func UploadPresignedUrl(key string, bucketname string) (string, error) {
	svc := s3.New(session.Must(session.NewSession()))

	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(key),
	})

	url, err := req.Presign(1 * time.Minute)
	if err != nil {
		return "", err
	}
	return url, nil
}

func ViewPresignedUrl(key string, bucketname string) (string, error) {
	svc := s3.New(session.Must(session.NewSession()))

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(key),
	})

	url, err := req.Presign(15 * time.Minute)
	if err != nil {
		return "", err
	}
	return url, nil
}
