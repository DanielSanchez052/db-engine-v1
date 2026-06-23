package heapfile

import (
	"db-engine-v1/internal/storage/page"
	"db-engine-v1/internal/storage/record"
	"slices"
)

func (h *HeapFile) getPageWithSpace(recordSize uint16) (*page.Page, bool, error) {
	for _, pageID := range h.metadata.PageIDs {
		page, err := h.pager.LoadPage(pageID)
		if err != nil {
			return nil, false, err
		}

		if page.CanFit(recordSize) {
			return page, false, nil
		}
	}

	newPage, err := h.allocatePage(page.DataPage) // revisar si debe quedar quemado o se debe recibir en parametro
	if err != nil {
		return nil, false, err
	}
	h.metadata.PageIDs = append(h.metadata.PageIDs, newPage.Header.PageID)
	return newPage, true, nil
}

func (h *HeapFile) GetRecord(rid *RecordID) (record.Record, error) {
	if rid.PageID == 0 {
		return nil, ErrInvalidRecordID
	}

	if !slices.Contains(h.metadata.PageIDs, rid.PageID) {
		return nil, ErrInvalidRecordID
	}

	page, err := h.pager.LoadPage(rid.PageID)
	if err != nil {
		return nil, err
	}
	return page.GetRecord(rid.SlotID)
}

func (h *HeapFile) Iterator() *Iterator {
	return NewIterator(h)
}
