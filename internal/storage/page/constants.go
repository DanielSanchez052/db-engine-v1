package page

const (
	PageIDOffset = 0
	PageIDSize   = 8

	RecordCountOffset = 8
	RecordCountSize   = 2

	FreeSpaceOffsetOffset = 10
	FreeSpaceOffsetSize   = 2

	SlotCountOffset = 12
	SlotCountSize   = 2

	PageTypeOffset = 14
	PageTypeSize   = 1

	ReservedOffset = 15
	ReservedSize   = 49
)
