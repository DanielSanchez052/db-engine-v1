package page

import (
	"db-engine-v1/internal/storage"
	"db-engine-v1/internal/storage/record"
	"db-engine-v1/internal/storage/slot"
)

type Page struct {
	Header  *PageHeader
	Payload [storage.PageSize - storage.PageHeaderSize]byte
}

func NewPage(id uint64, pageType PageType) *Page {
	return &Page{
		Header: NewPageHeader(id, pageType),
	}
}

func (p *Page) GetPageID() uint64 {
	return p.Header.PageID
}

func (p *Page) Serialize() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	headerBytes, err := p.Header.Serialize()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, storage.PageSize)
	copy(buf, headerBytes)
	copy(buf[storage.PageHeaderSize:], p.Payload[:])

	return buf, nil
}

func NewPageFromBytes(data []byte) (*Page, error) {
	if len(data) != storage.PageSize {
		return nil, ErrPageSizeMismatch
	}

	header, err := NewPageHeaderFromBytes(data[:storage.PageHeaderSize])
	if err != nil {
		return nil, err
	}

	page := &Page{
		Header:  header,
		Payload: [storage.PageSize - storage.PageHeaderSize]byte{},
	}

	copy(page.Payload[:], data[storage.PageHeaderSize:])

	if err := page.Validate(); err != nil {
		return nil, err
	}
	return page, nil
}

func (p *Page) Validate() error {
	var err error
	err = p.Header.Validate()
	if err != nil {
		return err
	}

	// Validate payload size
	if len(p.Payload) != storage.PageSize-storage.PageHeaderSize {
		return ErrPageSizeMismatch
	}

	return nil
}

func (p *Page) AvailableSpace() uint16 {
	usedTop := storage.PageHeaderSize + (p.Header.SlotCount * storage.SlotSize)
	return p.Header.FreeSpaceOffset - usedTop
}

func (p *Page) CanFit(recordSize uint16) bool {
	availableSpace := p.AvailableSpace()
	requiredSpace := recordSize + storage.SlotSize
	return requiredSpace <= availableSpace
}


