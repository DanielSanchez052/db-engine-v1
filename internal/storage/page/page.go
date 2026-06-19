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

func (p *Page) SlotOffset(slotID uint16) uint16 {
	return storage.PageHeaderSize + slotID*storage.SlotSize
}

func (p *Page) WriteSlot(slotID uint16, s *slot.Slot) error {
	if slotID > p.Header.SlotCount {
		return ErrInvalidSlotID
	}

	slotBytes, err := s.Serialize()
	if err != nil {
		return err
	}

	physicalOffset := p.SlotOffset(slotID)
	payloadOffset := physicalOffset - storage.PageHeaderSize

	end := payloadOffset + storage.SlotSize

	if end > uint16(len(p.Payload)) {
		return ErrSlotOutOfBounds
	}

	copy(
		p.Payload[payloadOffset:payloadOffset+storage.SlotSize],
		slotBytes,
	)
	return nil
}

func (p *Page) ReadSlot(slotID uint16) (*slot.Slot, error) {
	if slotID >= p.Header.SlotCount {
		return nil, ErrInvalidSlotID
	}

	physicalOffset := p.SlotOffset(slotID)
	payloadOffset := physicalOffset - storage.PageHeaderSize

	slotData := make([]byte, storage.SlotSize)
	copy(
		slotData,
		p.Payload[payloadOffset:payloadOffset+storage.SlotSize],
	)

	return slot.NewSlotFromBytes(slotData)
}

func (p *Page) InsertRecord(record record.Record) (uint16, error) {
	// aun no se tienen en cuenta los slots borrados esto llega en una fase posterior del proyecto

	if !p.CanFit(record.Size()) {
		return 0, ErrNotEnoughSpace
	}

	if record.Size() == 0 {
		return 0, ErrEmptyRecord
	}

	// es redundante pero lo hacemos como medida de proteccion
	if record.Size() > p.AvailableSpace() {
		return 0, ErrNotEnoughSpace
	}

	recordOffset := p.Header.FreeSpaceOffset - record.Size()

	err := p.writeRecord(recordOffset, record)
	if err != nil {
		return 0, err
	}

	newSlot := &slot.Slot{
		RecordOffset: recordOffset,
		RecordLength: record.Size(),
	}

	slotID := p.Header.SlotCount
	err = p.WriteSlot(slotID, newSlot)
	if err != nil {
		return 0, err
	}

	p.Header.RecordCount++
	p.Header.SlotCount++
	p.Header.FreeSpaceOffset = recordOffset

	return slotID, nil
}

func (p *Page) GetRecord(slotID uint16) (record.Record, error) {
	slot, err := p.ReadSlot(slotID)
	if err != nil {
		return nil, err
	}

	if slot.IsDeleted() {
		return nil, ErrRecordDeleted
	}

	return p.readRecord(slot.RecordOffset, slot.RecordLength)
}

func (p *Page) writeRecord(offset uint16, record record.Record) error {
	payloadOffset := offset - storage.PageHeaderSize

	end := payloadOffset + record.Size()

	if end > uint16(len(p.Payload)) {
		return ErrRecordOutOfBounds
	}

	copy(
		p.Payload[payloadOffset:end],
		record,
	)

	return nil
}

func (p *Page) readRecord(offset uint16, length uint16) (record.Record, error) {

	payloadOffset := offset - storage.PageHeaderSize

	end := payloadOffset + length

	if end > uint16(len(p.Payload)) {
		return nil, ErrRecordOutOfBounds
	}

	recordData := make([]byte, length)

	copy(
		recordData,
		p.Payload[payloadOffset:end],
	)

	return record.Record(recordData), nil
}

func (p *Page) DeleteRecord(slotID uint16) error {
	slot, err := p.ReadSlot(slotID)
	if err != nil {
		return err
	}

	if slot.IsDeleted() {
		return ErrRecordDeleted
	}

	slot.RecordLength = 0
	slot.RecordOffset = 0

	err = p.WriteSlot(slotID, slot)
	if err != nil {
		return err
	}

	p.Header.RecordCount--

	return nil
}
