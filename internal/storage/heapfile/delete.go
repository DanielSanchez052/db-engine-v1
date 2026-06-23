package heapfile



func (h *HeapFile) DeleteRecord(rid *RecordID) error {
	if rid == nil {
		return ErrInvalidRecordID
	}
	
	page, err := h.pager.LoadPage(rid.PageID)
	if err != nil {
		return err
	}
	
	err = page.DeleteRecord(rid.SlotID)
	if err != nil {
		return err
	}
	
	return h.pager.SavePage(page)
}