package record

import (
	"db-engine-v1/internal/storage"
)

const MaxRecordSize = storage.PageSize - storage.PageHeaderSize - storage.SlotSize

type Record []byte

func (r Record) Size() uint16 {
	return uint16(len(r))
}
