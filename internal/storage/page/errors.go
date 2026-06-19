package page

import "errors"

var (
	ErrPageHeaderSizeMismatch  = errors.New("page header data size mismatch")
	ErrInvalidPageType         = errors.New("invalid page type")
	ErrInvalidFreeSpaceOffset  = errors.New("free space offset exceeds page size")
	ErrFreeSpaceOffsetInHeader = errors.New("free space offset points within header")
	ErrPageSizeMismatch        = errors.New("page size mismatch")
	ErrInvalidSlotID           = errors.New("invalid slot ID")
	ErrSlotOutOfBounds         = errors.New("slot out of bounds")
	ErrNotEnoughSpace          = errors.New("not enough space")
	ErrEmptyRecord             = errors.New("empty record")
	ErrRecordDeleted           = errors.New("record deleted")
	ErrCorruptedSlot           = errors.New("corrupted slot")
	ErrRecordOutOfBounds       = errors.New("record out of bounds")
)
