package database

import "errors"

var (
	ErrHeaderSizeMismatch  = errors.New("header data size mismatch")
	ErrInvalidMagicNumber  = errors.New("invalid magic number")
	ErrInvalidPageSize     = errors.New("invalid page size")
	ErrInvalidPageChecksum = errors.New("invalid page checksum")
)
