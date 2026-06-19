package database

import (
	"db-engine-v1/internal/storage"
	"encoding/binary"
)

type DatabaseHeader struct {
	MagicNumber  [4]byte
	Version      uint16
	PageSize     uint16
	TotalPages   uint64
	FreePageHead uint64
	Reserved     [40]byte
}

func NewDatabaseHeader() *DatabaseHeader {
	return &DatabaseHeader{
		MagicNumber:  MagicNumber,
		Version:      Version,
		PageSize:     storage.PageSize,
		TotalPages:   1,
		FreePageHead: 0,
		Reserved:     [40]byte{},
	}
}

func (h *DatabaseHeader) Serialize() ([]byte, error) {
	if err := h.Validate(); err != nil {
		return nil, err
	}

	buf := make([]byte, storage.DatabaseHeaderSize)

	copy(buf[MagicNumberOffset:], h.MagicNumber[:])
	binary.LittleEndian.PutUint16(buf[VersionOffset:], h.Version)
	binary.LittleEndian.PutUint16(buf[PageSizeOffset:], h.PageSize)
	binary.LittleEndian.PutUint64(buf[TotalPagesOffset:], h.TotalPages)
	binary.LittleEndian.PutUint64(buf[FreePageHeadOffset:], h.FreePageHead)
	copy(buf[ReservedOffset:], h.Reserved[:])

	return buf, nil
}

func NewDatabaseHeaderFromBytes(data []byte) (*DatabaseHeader, error) {
	if len(data) != storage.DatabaseHeaderSize {
		return nil, ErrHeaderSizeMismatch
	}

	header := &DatabaseHeader{}

	copy(header.MagicNumber[:], data[MagicNumberOffset:MagicNumberOffset+MagicNumberSize])
	header.Version = binary.LittleEndian.Uint16(data[VersionOffset : VersionOffset+VersionSize])
	header.PageSize = binary.LittleEndian.Uint16(data[PageSizeOffset : PageSizeOffset+PageSizeSize])
	header.TotalPages = binary.LittleEndian.Uint64(data[TotalPagesOffset : TotalPagesOffset+TotalPagesSize])
	header.FreePageHead = binary.LittleEndian.Uint64(data[FreePageHeadOffset : FreePageHeadOffset+FreePageHeadSize])
	copy(header.Reserved[:], data[ReservedOffset:ReservedOffset+ReservedSize])

	if err := header.Validate(); err != nil {
		return nil, err
	}

	return header, nil
}

func (h *DatabaseHeader) Validate() error {
	if h.MagicNumber != MagicNumber {
		return ErrInvalidMagicNumber
	}

	if h.PageSize != storage.PageSize {
		return ErrInvalidPageSize
	}

	return nil
}
