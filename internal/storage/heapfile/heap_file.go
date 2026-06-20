package heapfile

import (
	"db-engine-v1/internal/catalog"
	"db-engine-v1/internal/storage/page"
	"db-engine-v1/internal/storage/pager"
	"db-engine-v1/internal/storage/record"
	"slices"
)

type HeapFile struct {
	pager        *pager.Pager
	metadata     *catalog.HeapMetadata
	allocatePage func(page.PageType) (*page.Page, error) // TODO: por el momento lo vamos a manejar asi, sin embargo luego se debe mover a otro lado
}

func New(pager *pager.Pager, metadata *catalog.HeapMetadata, allocatePage func(page.PageType) (*page.Page, error)) *HeapFile {
	return &HeapFile{
		pager:        pager,
		metadata:     metadata,
		allocatePage: allocatePage,
	}
}

func (h *HeapFile) InsertRecord(record record.Record) (*RecordID, error) {
	// buscar pagina con espacio
	recordSize := record.Size()

	page, err := h.getPageWithSpace(recordSize)

	if err != nil {
		return nil, err
	}

	slotID, err := page.InsertRecord(record)
	if err != nil {
		return nil, err
	}

	err = h.pager.SavePage(page)
	if err != nil {
		return nil, err
	}

	return &RecordID{
		PageID: page.Header.PageID,
		SlotID: slotID,
	}, nil
}

func (h *HeapFile) getPageWithSpace(recordSize uint16) (*page.Page, error) {
	for _, pageID := range h.metadata.PageIDs {
		page, err := h.pager.LoadPage(pageID)
		if err != nil {
			return nil, err
		}

		if page.CanFit(recordSize) {
			return page, nil
		}
	}

	newPage, err := h.allocatePage(page.DataPage) // revisar si debe quedar quemado o se debe recibir en parametro
	if err != nil {
		return nil, err
	}
	h.metadata.PageIDs = append(h.metadata.PageIDs, newPage.Header.PageID)
	return newPage, nil
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
