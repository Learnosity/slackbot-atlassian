package storage

import (
	"fmt"
	"io"

	"slackbot_atlassian/config"
	"slackbot_atlassian/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Client interface {
	PutObject(io.Reader, string) error

	GetFullURL(string) string
}

func New(cfg config.ResourceStorageConfig) Client {
	providers := []credentials.Provider{
		&credentials.StaticProvider{
			Value: credentials.Value{
				AccessKeyID:     cfg.AWS_Access_Key_ID,
				SecretAccessKey: cfg.AWS_Secret_Access_Key,
			},
		},
	}

	c := aws.NewConfig()

	c.WithRegion(cfg.S3_Region)
	c.WithCredentialsChainVerboseErrors(true)
	c.WithCredentials(credentials.NewChainCredentials(providers))

	session := session.New(c)

	return &client{
		cfg:      cfg,
		s3:       s3.New(session),
		uploader: s3manager.NewUploader(session),
	}
}

type client struct {
	cfg      config.ResourceStorageConfig
	s3       *s3.S3
	uploader *s3manager.Uploader
}

func (c *client) PutObject(rdr io.Reader, path string) error {
	log.LogF("Putting object to %s", path)
	acl := "public-read"
	_, err := c.uploader.Upload(&s3manager.UploadInput{
		ACL:    &acl,
		Bucket: &c.cfg.S3_Bucket,
		Key:    &path,
		Body:   rdr,
	})
	return err
}

func (c *client) GetFullURL(path string) string {
	return fmt.Sprintf("https://s3-%s.amazonaws.com/%s/%s", c.cfg.S3_Region, c.cfg.S3_Bucket, path)
}
