package slot

import "errors"

var (
	ErrSlotSizeMismatch    = errors.New("slot data size mismatch")
	ErrInvalidRecordOffset = errors.New("record offset exceeds page size")
	ErrInvalidRecordLength = errors.New("record length exceeds page size")
	ErrRecordExceedsPage   = errors.New("record offset + length exceeds page size")
)
