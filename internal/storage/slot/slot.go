package slot

import (
	"db-engine-v1/internal/storage"
	"encoding/binary"
)

type Slot struct {
	RecordOffset uint16
	RecordLength uint16
}

func (s *Slot) Serialize() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}

	buf := make([]byte, storage.SlotSize)

	binary.LittleEndian.PutUint16(buf[RecordOffsetOffset:], s.RecordOffset)
	binary.LittleEndian.PutUint16(buf[RecordLengthOffset:], s.RecordLength)

	return buf, nil
}

func NewSlotFromBytes(data []byte) (*Slot, error) {
	if len(data) != storage.SlotSize {
		return nil, ErrSlotSizeMismatch
	}

	slot := &Slot{}

	slot.RecordOffset = binary.LittleEndian.Uint16(data[RecordOffsetOffset : RecordOffsetOffset+RecordOffsetSize])
	slot.RecordLength = binary.LittleEndian.Uint16(data[RecordLengthOffset : RecordLengthOffset+RecordLengthSize])

	if err := slot.Validate(); err != nil {
		return nil, err
	}

	return slot, nil
}

func (s *Slot) IsDeleted() bool {
	return s.RecordOffset == 0 &&
		s.RecordLength == 0
}

func (s *Slot) Validate() error {
	// RecordOffset <= PageSize
	if uint32(s.RecordOffset) > storage.PageSize {
		return ErrInvalidRecordOffset
	}

	// RecordLength <= PageSize
	if uint32(s.RecordLength) > storage.PageSize {
		return ErrInvalidRecordLength
	}

	// RecordOffset + RecordLength <= PageSize
	if uint32(s.RecordOffset)+uint32(s.RecordLength) > storage.PageSize {
		return ErrRecordExceedsPage
	}

	return nil
}
