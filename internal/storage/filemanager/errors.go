package filemanager

import "errors"

var ErrShortRead = errors.New("short read")

var ErrShortWrite = errors.New("short write")
