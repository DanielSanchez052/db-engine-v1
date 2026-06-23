package heapfile

import "db-engine-v1/internal/storage/record"

func (h *HeapFile) InsertRecord(record record.Record) (*RecordID, bool, error) {
	// buscar pagina con espacio
	recordSize := record.Size()

	page, isNew, err := h.getPageWithSpace(recordSize)

	if err != nil {
		return nil, false, err
	}

	slotID, err := page.InsertRecord(record)
	if err != nil {
		return nil, false, err
	}

	err = h.pager.SavePage(page)
	if err != nil {
		return nil, false, err
	}

	return &RecordID{
		PageID: page.Header.PageID,
		SlotID: slotID,
	}, isNew, nil
}
