package r2

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/notion-echo/utils"
	"github.com/sirupsen/logrus"
)

var (
	bucketName      = os.Getenv(utils.BUCKET_NAME)
	bucketAccountId = os.Getenv(utils.BUCKET_ACCOUNT_ID)
	bucketAccessKey = os.Getenv(utils.BUCKET_ACCESS_KEY)
	bucketSecretKey = os.Getenv(utils.BUCKET_SECRET_KEY)
)

type R2Interface interface {
	UploadLogs(logFileName string, logger *logrus.Logger) error
}

type R2 struct {
	*s3.Client
}

func NewR2Client() (R2Interface, error) {
	fmt.Println("entering bucket setup:")
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", bucketAccountId),
		}, nil
	})
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(bucketAccessKey, bucketSecretKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}

	return &R2{
		Client: s3.NewFromConfig(cfg),
	}, nil
}

func (c R2) UploadLogs(logFileName string, logger *logrus.Logger) error {
	newLogFileName := fmt.Sprintf("logs-%s.log", time.Now().Format("2006-01-02"))
	err := os.Rename(logFileName, newLogFileName)
	if err != nil {
		return err
	}

	compressedLogFileName := newLogFileName + ".gz"
	err = utils.CompressFile(newLogFileName, compressedLogFileName)
	if err != nil {
		return err
	}

	err = c.uploadLogFileToR2(compressedLogFileName)
	if err != nil {
		return err
	}
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	logger.SetOutput(logFile)
	return nil
}

func (c R2) uploadLogFileToR2(logFileName string) error {
	ctx := context.Background()

	logFile, err := os.Open(logFileName)
	if err != nil {
		return err
	}
	defer logFile.Close()

	fileInfo, err := logFile.Stat()
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()
	buffer := make([]byte, fileSize)
	_, err = logFile.Read(buffer)
	if err != nil {
		return err
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(logFileName),
		Body:        bytes.NewReader(buffer),
		ContentType: aws.String("text/plain"),
	}

	_, err = c.Client.PutObject(ctx, input)
	if err != nil {
		return err
	}
	return nil
}
