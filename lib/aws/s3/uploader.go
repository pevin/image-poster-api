package s3

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func GetS3Uploader() *s3manager.Uploader {
	awsRegion := os.Getenv("APP_AWS_REGION")
	conf := &aws.Config{Region: aws.String(awsRegion)}
	sess, err := session.NewSession(conf)
	if err != nil {
		fmt.Println("Error creating aws session: ", err)
		panic(err)
	}
	uploader := s3manager.NewUploader(sess)
	return uploader
}
