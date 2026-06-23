package page

import (
	"db-engine-v1/internal/storage"
	"db-engine-v1/internal/storage/record"
	"db-engine-v1/internal/storage/slot"
)

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

	if p.Header.RecordCount > 0 {
		p.Header.RecordCount--
	}

	return nil
}