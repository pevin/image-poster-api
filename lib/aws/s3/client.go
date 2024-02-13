package s3

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func GetS3Client() *s3.S3 {
	awsRegion := os.Getenv("APP_AWS_REGION")
	conf := &aws.Config{Region: aws.String(awsRegion)}
	sess, err := session.NewSession(conf)
	if err != nil {
		fmt.Println("Error creating aws session: ", err)
		panic(err)
	}
	return s3.New(sess, &aws.Config{
		Region: aws.String(awsRegion),
	})

}
