package page

import (
	"db-engine-v1/internal/storage"
	"db-engine-v1/internal/storage/slot"
)

func (p *Page) SlotOffset(slotID uint16) uint16 {
	return storage.PageHeaderSize + slotID*storage.SlotSize
}

func (p *Page) GetSlotCount() uint16 {
	return p.Header.SlotCount
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