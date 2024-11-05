// internal/storage/s3.go

package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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

func NewS3Storage(region, endpoint, accessKeyID, secretAccessKey string, logger *logger.Logger) (*S3Storage, error) {
	logger.Debug("Initializing S3 storage with region: %s, endpoint: %s", region, endpoint)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{URL: endpoint}, nil
			})),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	)
	if err != nil {
		logger.Error("Failed to load S3 config: %v", err)
		return nil, fmt.Errorf("failed to load S3 config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &S3Storage{
		client: client,
		logger: logger,
	}, nil
}

func (s *S3Storage) Read(path string) ([]byte, error) {
	s.logger.Debug("Reading from S3: %s", path)
	bucket, key, err := ParseS3URI(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse S3 URI: %w", err)
	}

	s.logger.Debug("Parsed S3 path - Bucket: %s, Key: %s", bucket, key)

	result, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		s.logger.Error("Failed to read file from S3: %v", err)
		return nil, fmt.Errorf("failed to read file from S3: %w", err)
	}
	defer result.Body.Close()

	content, err := ioutil.ReadAll(result.Body)
	if err != nil {
		s.logger.Error("Failed to read content from S3 object: %v", err)
		return nil, fmt.Errorf("failed to read content from S3 object: %w", err)
	}

	s.logger.Debug("Successfully read %d bytes from S3", len(content))
	return content, nil
}

func (s *S3Storage) Write(path string, data []byte) error {
	s.logger.Debug("Writing file to S3: %s", path)

	bucket, key, err := ParseS3URI(path)
	if err != nil {
		return fmt.Errorf("failed to parse S3 URI: %w", err)
	}

	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		s.logger.Error("Failed to write file to S3: %v", err)
		return fmt.Errorf("failed to write file to S3: %w", err)
	}

	s.logger.Debug("Successfully wrote %d bytes to S3", len(data))
	return nil
}

func (s *S3Storage) List(prefix string) ([]string, error) {
	s.logger.Debug("Listing files in S3 with prefix: %s", prefix)

	bucket, key, err := ParseS3URI(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to parse S3 URI: %w", err)
	}

	var files []string
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(key),
	})

	// Extrayez le domaine de l'URL originale
	domainParts := strings.SplitN(prefix, "/", 4)
	domain := strings.Join(domainParts[:3], "/")

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			s.logger.Error("Failed to list files in S3: %v", err)
			return nil, fmt.Errorf("failed to list files in S3: %w", err)
		}
		for _, obj := range page.Contents {
			if !strings.HasSuffix(*obj.Key, "/") { // Ignore directory markers
				files = append(files, fmt.Sprintf("%s/%s/%s", domain, bucket, *obj.Key))
			}
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

	bucket, key, err := ParseS3URI(path)
	if err != nil {
		return false, fmt.Errorf("failed to parse S3 URI: %w", err)
	}
	s.logger.Error("Parsed S3 URI: %s %s", bucket, key)

	// Assurez-vous que la clé se termine par '/'
	if !strings.HasSuffix(key, "/") {
		key += "/"
	}

	s.logger.Debug("Checking key %s on bucket %s for path %s", key, bucket, path)

	result, err := s.client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Prefix:    aws.String(key),
		Delimiter: aws.String("/"),
		MaxKeys:   aws.Int32(1),
	})
	if err != nil {
		s.logger.Error("Error checking if path is a directory in S3: %v", err)
		return false, fmt.Errorf("error checking if path is a directory in S3: %w", err)
	}

	return len(result.CommonPrefixes) > 0 || len(result.Contents) > 0, nil
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

func (s *S3Storage) ReadFromBucket(bucket, key string) ([]byte, error) {
	s.logger.Debug("Reading file from S3: bucket=%s, key=%s", bucket, key)

	result, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		s.logger.Error("Failed to read file from S3: %v", err)
		return nil, fmt.Errorf("failed to read file from S3: %w", err)
	}
	defer result.Body.Close()

	return ioutil.ReadAll(result.Body)
}

func (s *S3Storage) StatObject(bucket, key string) (os.FileInfo, error) {
	s.logger.Debug("Getting file info from S3 - Bucket: %s, Key: %s", bucket, key)

	result, err := s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		s.logger.Error("Failed to get file info from S3: %v", err)
		return nil, fmt.Errorf("failed to get file info from S3: %w", err)
	}

	return &s3FileInfo{
		name:    filepath.Base(key),
		size:    *result.ContentLength,
		modTime: *result.LastModified,
	}, nil
}

func ParseS3URI(uri string) (bucket, key string, err error) {
	if !strings.HasPrefix(strings.ToLower(uri), "s3://") {
		return "", "", fmt.Errorf("invalid S3 URI format on prefix")
	}

	// Remove the "s3://" prefix
	uri = strings.TrimPrefix(uri, "s3://")

	// Split the remaining string into parts
	parts := strings.SplitN(uri, "/", 3)

	if len(parts) < 3 {
		return "", "", fmt.Errorf("invalid S3 URI format on len(part) : %s", uri)
	}

	// The first part is the domain
	// The second part is the bucket
	bucket = parts[1]
	// The key is everything after the bucket
	key = parts[2]

	return bucket, key, nil
}

func (s *S3Storage) readS3Directory(bucket, prefix string) ([]byte, error) {
	s.logger.Debug("Reading S3 directory: bucket=%s, prefix=%s", bucket, prefix)
	var content []byte
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to list S3 objects: %w", err)
		}

		for _, obj := range page.Contents {
			if !strings.HasSuffix(*obj.Key, "/") { // Skip directory markers
				objContent, err := s.getS3ObjectContent(bucket, *obj.Key)
				if err != nil {
					return nil, err
				}
				content = append(content, objContent...)
				content = append(content, '\n') // Add newline between files
			}
		}
	}

	if len(content) == 0 {
		return nil, fmt.Errorf("no content found in S3 directory: %s/%s", bucket, prefix)
	}

	s.logger.Debug("Successfully read %d bytes from S3 directory", len(content))
	return content, nil
}

func (s *S3Storage) getS3ObjectContent(bucket, key string) ([]byte, error) {
	result, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 object: %w", err)
	}
	defer result.Body.Close()

	return ioutil.ReadAll(result.Body)
}

func (s *S3Storage) GetReader(path string) (io.ReadCloser, error) {
	s.logger.Debug("Getting reader for S3 object: %s", path)
	bucket, key, err := ParseS3URI(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse S3 URI: %w", err)
	}

	result, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 object: %w", err)
	}

	return result.Body, nil
}
