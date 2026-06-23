package database

import "errors"

var (
	ErrHeaderSizeMismatch  = errors.New("header data size mismatch")
	ErrInvalidMagicNumber  = errors.New("invalid magic number")
	ErrInvalidPageSize     = errors.New("invalid page size")
	ErrInvalidPageChecksum = errors.New("invalid page checksum")
	ErrInvalidDatabase     = errors.New("invalid database")
	ErrDatabaseExists      = errors.New("database already exists")
	ErrInvalidTableName    = errors.New("invalid table name")
	ErrInvalidColumns      = errors.New("invalid columns")
	ErrTableNotFound       = errors.New("table not found")
	ErrTableAlreadyExists  = errors.New("table already exists")
	ErrInvalidRecord       = errors.New("invalid record")
	ErrInvalidRecordID     = errors.New("invalid record id")
)
