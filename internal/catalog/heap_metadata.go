package catalog

import (
	"encoding/binary"
	"errors"
)

var (
	ErrInvalidMetadata = errors.New("invalid metadata")
	ErrHeapNameTooLong = errors.New("heap name too long")
)

type HeapMetadata struct {
	Name    string
	PageIDs []uint64
}

func (h *HeapMetadata) Serialize() ([]byte, error) {
	nameLength := uint16(len(h.Name))
	pageCount := uint32(len(h.PageIDs))

	if nameLength > MaxHeapNameLength {
		return nil, ErrHeapNameTooLong
	}

	buffer := make([]byte, h.Size())
	offset := 0

	if err := writeString(buffer, &offset, h.Name); err != nil {
		return nil, ErrInvalidMetadata
	}

	binary.LittleEndian.PutUint32(buffer[offset:offset+PageCountSize], pageCount)
	offset += PageCountSize

	for _, pageID := range h.PageIDs {
		binary.LittleEndian.PutUint64(buffer[offset:offset+PageIDSize], pageID)
		offset += PageIDSize
	}

	return buffer, nil
}

func NewHeapMetadataFromBytes(data []byte) (*HeapMetadata, error) {
	if len(data) < HeapNameLengthSize {
		return nil, ErrInvalidMetadata
	}

	offset := 0
	name, err := readString(data, &offset)
	if err != nil {
		return nil, ErrInvalidMetadata
	}

	if offset+PageCountSize > len(data) {
		return nil, ErrInvalidMetadata
	}

	pageCount := binary.LittleEndian.Uint32(data[offset : offset+PageCountSize])
	offset += PageCountSize

	expected := offset + int(pageCount)*PageIDSize

	if len(data) < expected {
		return nil, ErrInvalidMetadata
	}

	pageIDs := make([]uint64, int(pageCount))
	for i := 0; i < int(pageCount); i++ {
		pageIDs[i] = binary.LittleEndian.Uint64(data[offset : offset+PageIDSize])
		offset += PageIDSize
	}

	return &HeapMetadata{
		Name:    name,
		PageIDs: pageIDs,
	}, nil
}

func (h *HeapMetadata) Size() int {
	return HeapNameLengthSize + len(h.Name) + PageCountSize + (len(h.PageIDs) * PageIDSize)
}

func (h *HeapMetadata) AddPage(pageID uint64) {
	h.PageIDs = append(h.PageIDs, pageID)
}
