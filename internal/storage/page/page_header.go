package page

import (
	"db-engine-v1/internal/storage"
	"encoding/binary"
)

type PageHeader struct {
	PageID          uint64
	RecordCount     uint16
	FreeSpaceOffset uint16
	SlotCount       uint16
	PageType        PageType
	Reserved        [49]byte
}

func NewPageHeader(id uint64, pageType PageType) *PageHeader {
	return &PageHeader{
		PageID:          id,
		PageType:        pageType,
		RecordCount:     0,
		FreeSpaceOffset: storage.PageSize,
		SlotCount:       0,
		Reserved:        [49]byte{},
	}
}

func (p *PageHeader) Serialize() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	buf := make([]byte, PageIDSize+RecordCountSize+FreeSpaceOffsetSize+SlotCountSize+PageTypeSize+ReservedSize)

	binary.LittleEndian.PutUint64(buf[PageIDOffset:], p.PageID)
	binary.LittleEndian.PutUint16(buf[RecordCountOffset:], p.RecordCount)
	binary.LittleEndian.PutUint16(buf[FreeSpaceOffsetOffset:], p.FreeSpaceOffset)
	binary.LittleEndian.PutUint16(buf[SlotCountOffset:], p.SlotCount)
	buf[PageTypeOffset] = byte(p.PageType)
	copy(buf[ReservedOffset:], p.Reserved[:])

	return buf, nil
}

func NewPageHeaderFromBytes(data []byte) (*PageHeader, error) {
	expectedSize := storage.PageHeaderSize
	if len(data) != expectedSize {
		return nil, ErrPageHeaderSizeMismatch
	}

	header := &PageHeader{}

	header.PageID = binary.LittleEndian.Uint64(data[PageIDOffset : PageIDOffset+PageIDSize])
	header.RecordCount = binary.LittleEndian.Uint16(data[RecordCountOffset : RecordCountOffset+RecordCountSize])
	header.FreeSpaceOffset = binary.LittleEndian.Uint16(data[FreeSpaceOffsetOffset : FreeSpaceOffsetOffset+FreeSpaceOffsetSize])
	header.SlotCount = binary.LittleEndian.Uint16(data[SlotCountOffset : SlotCountOffset+SlotCountSize])
	header.PageType = PageType(data[PageTypeOffset])
	copy(header.Reserved[:], data[ReservedOffset:ReservedOffset+ReservedSize])

	if err := header.Validate(); err != nil {
		return nil, err
	}

	return header, nil
}

func (p *PageHeader) Validate() error {
	// Valid PageType check
	if !p.PageType.IsValid() {
		return ErrInvalidPageType
	}

	// FreeSpaceOffset must not exceed page size
	if p.FreeSpaceOffset > storage.PageSize {
		return ErrInvalidFreeSpaceOffset
	}

	// FreeSpaceOffset must not point within the header
	if p.FreeSpaceOffset < storage.PageHeaderSize {
		return ErrFreeSpaceOffsetInHeader
	}

	return nil
}
