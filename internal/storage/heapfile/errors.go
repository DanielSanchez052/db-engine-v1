package heapfile

import "errors"

var (
	ErrInvalidRecordID = errors.New("invalid record id")
	ErrIteratorDone    = errors.New("iterator done")
)
