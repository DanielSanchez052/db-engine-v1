package database

const (
	DatabaseHeaderSize = 64

	MagicNumberOffset = 0
	MagicNumberSize   = 4

	VersionOffset = 4
	VersionSize   = 2

	PageSizeOffset = 6
	PageSizeSize   = 2

	TotalPagesOffset = 8
	TotalPagesSize   = 8

	FreePageHeadOffset = 16
	FreePageHeadSize   = 8

	ReservedOffset = 24
	ReservedSize   = 49
)

const PageSize = 4096

var MagicNumber = [4]byte{'M', 'N', 'D', 'B'}
var Version = uint16(1)
