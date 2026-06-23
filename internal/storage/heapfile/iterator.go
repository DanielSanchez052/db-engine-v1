package heapfile

import (
	"db-engine-v1/internal/storage/page"
	"db-engine-v1/internal/storage/record"
	"errors"
)

type Iterator struct {
	heap *HeapFile

	currentPageIndex int
	currentSlotID    uint16

	currentPage *page.Page
}

func NewIterator(heap *HeapFile) *Iterator {
	return &Iterator{
		heap:             heap,
		currentPageIndex: 0,
		currentSlotID:    0,
	}
}

func (it *Iterator) loadCurrentPage() error {
	if it.currentPageIndex >= len(it.heap.metadata.PageIDs) {
		return nil
	}

	pageID := it.heap.metadata.PageIDs[it.currentPageIndex]
	page, err := it.heap.pager.LoadPage(pageID)
	if err != nil {
		return err
	}

	it.currentPage = page
	it.currentSlotID = 0

	return nil
}

func (it *Iterator) Next() (record.Record, *RecordID, error) {
	for {
		if it.currentPage == nil {
			if err := it.loadCurrentPage(); err != nil {
				return nil, nil, err
			}
		}

		if it.currentSlotID >= it.currentPage.GetSlotCount() {

			it.currentPageIndex++
			it.currentSlotID = 0

			if it.currentPageIndex >= len(it.heap.metadata.PageIDs) {
				return nil, nil, ErrIteratorDone
			}

			it.currentPage = nil
			continue
		}

		slotID := it.currentSlotID
		it.currentSlotID++

		rec, err := it.currentPage.GetRecord(slotID)

		if errors.Is(err, page.ErrRecordDeleted) {
			continue
		}

		if err != nil {
			return nil, nil, err
		}

		return rec, NewRecordID(it.currentPage.GetPageID(), slotID), nil
	}
}
