// internal/storage/s3_test.go

package storage

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogger est un mock de l'interface Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Info(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Warning(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Error(format string, args ...interface{}) {
	m.Called(format, args)
}

// MockS3Client est un mock du client S3
type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

func (m *MockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func (m *MockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.ListObjectsV2Output), args.Error(1)
}

func (m *MockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.DeleteObjectOutput), args.Error(1)
}

func (m *MockS3Client) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.HeadObjectOutput), args.Error(1)
}

func TestS3StorageRead(t *testing.T) {
	mockClient := new(MockS3Client)
	mockLogger := new(MockLogger)
	s3Storage := &S3Storage{
		client: mockClient,
		bucket: "test-bucket",
		logger: mockLogger,
	}

	expectedContent := []byte("test content")
	mockClient.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(
		&s3.GetObjectOutput{
			Body: ioutil.NopCloser(bytes.NewReader(expectedContent)),
		},
		nil,
	)

	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	content, err := s3Storage.Read("test.txt")

	assert.NoError(t, err)
	assert.Equal(t, expectedContent, content)
	mockClient.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestS3StorageWrite(t *testing.T) {
	mockClient := new(MockS3Client)
	mockLogger := new(MockLogger)
	s3Storage := &S3Storage{
		client: mockClient,
		bucket: "test-bucket",
		logger: mockLogger,
	}

	content := []byte("test content")
	mockClient.On("PutObject", mock.Anything, mock.Anything, mock.Anything).Return(
		&s3.PutObjectOutput{},
		nil,
	)

	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	err := s3Storage.Write("test.txt", content)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestS3StorageList(t *testing.T) {
	mockClient := new(MockS3Client)
	mockLogger := new(MockLogger)
	s3Storage := &S3Storage{
		client: mockClient,
		bucket: "test-bucket",
		logger: mockLogger,
	}

	mockClient.On("ListObjectsV2", mock.Anything, mock.Anything, mock.Anything).Return(
		&s3.ListObjectsV2Output{
			Contents: []types.Object{
				{Key: aws.String("file1.txt")},
				{Key: aws.String("file2.txt")},
			},
		},
		nil,
	)

	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	files, err := s3Storage.List("prefix")

	assert.NoError(t, err)
	assert.Equal(t, []string{"file1.txt", "file2.txt"}, files)
	mockClient.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestS3StorageDelete(t *testing.T) {
	mockClient := new(MockS3Client)
	mockLogger := new(MockLogger)
	s3Storage := &S3Storage{
		client: mockClient,
		bucket: "test-bucket",
		logger: mockLogger,
	}

	mockClient.On("DeleteObject", mock.Anything, mock.Anything, mock.Anything).Return(
		&s3.DeleteObjectOutput{},
		nil,
	)

	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	err := s3Storage.Delete("test.txt")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestS3StorageExists(t *testing.T) {
	mockClient := new(MockS3Client)
	mockLogger := new(MockLogger)
	s3Storage := &S3Storage{
		client: mockClient,
		bucket: "test-bucket",
		logger: mockLogger,
	}

	mockClient.On("HeadObject", mock.Anything, mock.Anything, mock.Anything).Return(
		&s3.HeadObjectOutput{},
		nil,
	)

	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	exists, err := s3Storage.Exists("test.txt")

	assert.NoError(t, err)
	assert.True(t, exists)
	mockClient.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestS3StorageIsDirectory(t *testing.T) {
	mockClient := new(MockS3Client)
	mockLogger := new(MockLogger)
	s3Storage := &S3Storage{
		client: mockClient,
		bucket: "test-bucket",
		logger: mockLogger,
	}

	mockClient.On("ListObjectsV2", mock.Anything, mock.Anything, mock.Anything).Return(
		&s3.ListObjectsV2Output{
			CommonPrefixes: []types.CommonPrefix{
				{Prefix: aws.String("test/")},
			},
		},
		nil,
	)

	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	isDir, err := s3Storage.IsDirectory("test/")

	assert.NoError(t, err)
	assert.True(t, isDir)
	mockClient.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestS3StorageStat(t *testing.T) {
	mockClient := new(MockS3Client)
	mockLogger := new(MockLogger)
	s3Storage := &S3Storage{
		client: mockClient,
		bucket: "test-bucket",
		logger: mockLogger,
	}

	lastModified := time.Now()
	mockClient.On("HeadObject", mock.Anything, mock.Anything, mock.Anything).Return(
		&s3.HeadObjectOutput{
			ContentLength: aws.Int64(100),
			LastModified:  &lastModified,
		},
		nil,
	)

	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	info, err := s3Storage.Stat("test.txt")

	assert.NoError(t, err)
	assert.Equal(t, "test.txt", info.Name())
	assert.Equal(t, int64(100), info.Size())
	assert.Equal(t, lastModified, info.ModTime())
	mockClient.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
