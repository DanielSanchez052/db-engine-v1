package page

type Page struct {
	Header  *PageHeader
	Payload [PageSize - PageHeaderSize]byte
}

func NewPage(id uint64, pageType PageType) *Page {
	return &Page{
		Header: NewPageHeader(id, pageType),
	}
}

func (p *Page) Serialize() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	headerBytes, err := p.Header.Serialize()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, PageSize)
	copy(buf, headerBytes)
	copy(buf[PageHeaderSize:], p.Payload[:])

	return buf, nil
}

func NewPageFromBytes(data []byte) (*Page, error) {
	if len(data) != PageSize {
		return nil, ErrPageSizeMismatch
	}

	header, err := NewPageHeaderFromBytes(data[:PageHeaderSize])
	if err != nil {
		return nil, err
	}

	page := &Page{
		Header:  header,
		Payload: [PageSize - PageHeaderSize]byte{},
	}

	copy(page.Payload[:], data[PageHeaderSize:])

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
	if len(p.Payload) != PageSize-PageHeaderSize {
		return ErrPageSizeMismatch
	}

	return nil
}

func (p *Page) FreeSpace() uint16 {
	return p.Header.FreeSpaceOffset
}
