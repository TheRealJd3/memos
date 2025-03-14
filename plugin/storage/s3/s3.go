package s3

import (
	"context"
	"fmt"
	"io"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Config struct {
	AccessKey string
	SecretKey string
	Bucket    string
	EndPoint  string
	Path      string
	Region    string
	URLPrefix string
}

type Client struct {
	Client *awss3.Client
	Config *Config
}

func NewClient(ctx context.Context, config *Config) (*Client, error) {
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           config.EndPoint,
			SigningRegion: config.Region,
		}, nil
	})

	awsConfig, err := s3config.LoadDefaultConfig(ctx,
		s3config.WithEndpointResolverWithOptions(resolver),
		s3config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.AccessKey, config.SecretKey, "")),
	)
	if err != nil {
		return nil, err
	}

	client := awss3.NewFromConfig(awsConfig)

	return &Client{
		Client: client,
		Config: config,
	}, nil
}

func (client *Client) UploadFile(ctx context.Context, filename string, fileType string, src io.Reader) (string, error) {
	uploader := manager.NewUploader(client.Client)
	uploadOutput, err := uploader.Upload(ctx, &awss3.PutObjectInput{
		Bucket:      aws.String(client.Config.Bucket),
		Key:         aws.String(path.Join(client.Config.Path, filename)),
		Body:        src,
		ContentType: aws.String(fileType),
		ACL:         types.ObjectCannedACL(*aws.String("public-read")),
	})
	if err != nil {
		return "", err
	}

	link := uploadOutput.Location
	// If url prefix is set, use it as the file link.
	if client.Config.URLPrefix != "" {
		link = fmt.Sprintf("%s/%s", client.Config.URLPrefix, filename)
	}
	if link == "" {
		return "", fmt.Errorf("failed to get file link")
	}
	return link, nil
}
