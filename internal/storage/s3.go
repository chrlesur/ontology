// internal/storage/s3.go

package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/chrlesur/Ontology/internal/logger"
)

type S3ClientInterface interface {
    GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
    PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
    ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
    DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
    HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
}

type S3Storage struct {
    client S3ClientInterface
    bucket string
    logger Logger
}

func NewS3Storage(bucket, region, endpoint, accessKeyID, secretAccessKey string, logger *logger.Logger) (*S3Storage, error) {
	logger.Debug("Initializing S3 storage with bucket: %s, region: %s, endpoint: %s", bucket, region, endpoint)

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, reg string, opts ...interface{}) (aws.Endpoint, error) {
		if endpoint != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           endpoint,
				SigningRegion: region,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	)
	if err != nil {
		logger.Error("Failed to load S3 config: %v", err)
		return nil, fmt.Errorf("failed to load S3 config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Storage{
        client: client,
        bucket: bucket,
        logger: logger,
    }, nil
}

func (s *S3Storage) Read(path string) ([]byte, error) {
	s.logger.Debug("Reading file from S3: %s", path)

	result, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		s.logger.Error("Failed to read file from S3: %v", err)
		return nil, fmt.Errorf("failed to read file from S3: %w", err)
	}
	defer result.Body.Close()

	return ioutil.ReadAll(result.Body)
}

func (s *S3Storage) Write(path string, data []byte) error {
	s.logger.Debug("Writing file to S3: %s", path)

	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		s.logger.Error("Failed to write file to S3: %v", err)
		return fmt.Errorf("failed to write file to S3: %w", err)
	}

	return nil
}

func (s *S3Storage) List(prefix string) ([]string, error) {
	s.logger.Debug("Listing files in S3 with prefix: %s", prefix)

	var files []string
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			s.logger.Error("Failed to list files in S3: %v", err)
			return nil, fmt.Errorf("failed to list files in S3: %w", err)
		}
		for _, obj := range page.Contents {
			files = append(files, *obj.Key)
		}
	}

	return files, nil
}

func (s *S3Storage) Delete(path string) error {
	s.logger.Debug("Deleting file from S3: %s", path)

	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		s.logger.Error("Failed to delete file from S3: %v", err)
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}

func (s *S3Storage) Exists(path string) (bool, error) {
	s.logger.Debug("Checking if file exists in S3: %s", path)

	_, err := s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) || strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		s.logger.Error("Error checking if file exists in S3: %v", err)
		return false, fmt.Errorf("error checking if file exists in S3: %w", err)
	}

	return true, nil
}

func (s *S3Storage) IsDirectory(path string) (bool, error) {
	s.logger.Debug("Checking if path is a directory in S3: %s", path)

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	result, err := s.client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket:    aws.String(s.bucket),
		Prefix:    aws.String(path),
		Delimiter: aws.String("/"),
		MaxKeys:   aws.Int32(1),
	})
	if err != nil {
		s.logger.Error("Error checking if path is a directory in S3: %v", err)
		return false, fmt.Errorf("error checking if path is a directory in S3: %w", err)
	}

	return len(result.CommonPrefixes) > 0 || (len(result.Contents) > 0 && *result.Contents[0].Key == path), nil
}

func (s *S3Storage) Stat(path string) (FileInfo, error) {
	s.logger.Debug("Getting file info from S3: %s", path)

	result, err := s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		s.logger.Error("Failed to get file info from S3: %v", err)
		return nil, fmt.Errorf("failed to get file info from S3: %w", err)
	}

	size := result.ContentLength
	if size == nil {
		size = aws.Int64(0)
	}
	return &s3FileInfo{
		name:    filepath.Base(path),
		size:    *size,
		modTime: *result.LastModified,
	}, nil
}

type s3FileInfo struct {
	name    string
	size    int64
	modTime time.Time
}

func (fi *s3FileInfo) Name() string       { return fi.name }
func (fi *s3FileInfo) Size() int64        { return fi.size }
func (fi *s3FileInfo) Mode() os.FileMode  { return 0 }
func (fi *s3FileInfo) ModTime() time.Time { return fi.modTime }
func (fi *s3FileInfo) IsDir() bool        { return false }
func (fi *s3FileInfo) Sys() interface{}   { return nil }
