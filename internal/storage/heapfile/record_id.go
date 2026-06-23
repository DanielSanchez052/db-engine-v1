package heapfile

type RecordID struct {
	PageID uint64
	SlotID uint16
}

func NewRecordID(pageID uint64, slotID uint16) *RecordID {
	return &RecordID{
		PageID: pageID,
		SlotID: slotID,
	}
}
