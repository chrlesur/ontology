// internal/storage/errors.go

package storage

import "errors"

var (
    // ErrInvalidS3URI est renvoyée lorsqu'une URI S3 est invalide
    ErrInvalidS3URI = errors.New("invalid S3 URI format")
)