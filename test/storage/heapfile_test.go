package storage_test

import (
	"testing"

	"db-engine-v1/internal/catalog"
	"db-engine-v1/internal/storage/filemanager"
	"db-engine-v1/internal/storage/heapfile"
	"db-engine-v1/internal/storage/page"
	"db-engine-v1/internal/storage/pager"
	"db-engine-v1/internal/storage/record"
)

func setupHeapFile(t *testing.T) (*heapfile.HeapFile, *pager.Pager, *catalog.HeapMetadata, func()) {
	dir := t.TempDir()
	path := dir + "/test.db"

	fm, err := filemanager.Open(path)
	if err != nil {
		t.Fatalf("filemanager.Open() error = %v", err)
	}

	pg := pager.NewPager(fm)
	metadata := &catalog.HeapMetadata{
		Name:    "test_heap",
		PageIDs: []uint64{},
	}

	nextPageID := uint64(1)
	allocatePage := func(pageType page.PageType) (*page.Page, error) {
		newPage := page.NewPage(nextPageID, pageType)
		nextPageID++
		err := pg.SavePage(newPage)
		return newPage, err
	}

	hf := heapfile.New(pg, metadata, allocatePage)

	cleanup := func() {
		fm.Close()
	}

	return hf, pg, metadata, cleanup
}

func TestHeapFileNew(t *testing.T) {
	hf, pg, metadata, cleanup := setupHeapFile(t)
	defer cleanup()

	if hf == nil {
		t.Fatalf("New() returned nil")
	}

	_ = pg
	_ = metadata
}

func TestHeapFileInsertRecord(t *testing.T) {
	hf, _, _, cleanup := setupHeapFile(t)
	defer cleanup()

	rec := record.Record([]byte("hello"))

	rid, err := hf.InsertRecord(rec)
	if err != nil {
		t.Fatalf("InsertRecord() error = %v, want nil", err)
	}

	if rid == nil {
		t.Fatalf("InsertRecord() returned nil RecordID")
	}

	if rid.PageID != 1 {
		t.Errorf("PageID = %d, want 1", rid.PageID)
	}

	if rid.SlotID != 0 {
		t.Errorf("SlotID = %d, want 0", rid.SlotID)
	}
}

func TestHeapFileInsertAndGetRecord(t *testing.T) {
	hf, _, _, cleanup := setupHeapFile(t)
	defer cleanup()

	rec := record.Record([]byte("hello world"))

	rid, err := hf.InsertRecord(rec)
	if err != nil {
		t.Fatalf("InsertRecord() error = %v", err)
	}

	got, err := hf.GetRecord(rid)
	if err != nil {
		t.Fatalf("GetRecord() error = %v", err)
	}

	if string(got) != string(rec) {
		t.Errorf("GetRecord() = %q, want %q", string(got), string(rec))
	}
}

func TestHeapFileInsertMultipleRecords(t *testing.T) {
	hf, _, _, cleanup := setupHeapFile(t)
	defer cleanup()

	records := []record.Record{
		record.Record([]byte("first")),
		record.Record([]byte("second")),
		record.Record([]byte("third")),
	}

	var rids []*heapfile.RecordID
	for _, rec := range records {
		rid, err := hf.InsertRecord(rec)
		if err != nil {
			t.Fatalf("InsertRecord(%q) error = %v", string(rec), err)
		}
		rids = append(rids, rid)
	}

	for i, rid := range rids {
		got, err := hf.GetRecord(rid)
		if err != nil {
			t.Fatalf("GetRecord(%d) error = %v", i, err)
		}
		if string(got) != string(records[i]) {
			t.Errorf("GetRecord(%d) = %q, want %q", i, string(got), string(records[i]))
		}
	}
}

func TestHeapFileInsertFillsPage(t *testing.T) {
	hf, _, metadata, cleanup := setupHeapFile(t)
	defer cleanup()

	// Insert large records to force allocation of subsequent pages
	largeData := make([]byte, 3000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	for i := 0; i < 5; i++ {
		rec := record.Record(largeData)
		_, err := hf.InsertRecord(rec)
		if err != nil {
			t.Fatalf("InsertRecord(%d) error = %v", i, err)
		}
	}

	if len(metadata.PageIDs) < 2 {
		t.Errorf("len(PageIDs) = %d, want multiple pages (3000*5 bytes needs >1 page)", len(metadata.PageIDs))
	}
}

func TestHeapFileGetRecordInvalidPageID(t *testing.T) {
	hf, _, _, cleanup := setupHeapFile(t)
	defer cleanup()

	rid := &heapfile.RecordID{PageID: 0, SlotID: 0}
	_, err := hf.GetRecord(rid)
	if err == nil {
		t.Errorf("GetRecord() expected error for PageID=0, got nil")
	}
}

func TestHeapFileGetRecordNonExistentPage(t *testing.T) {
	hf, _, _, cleanup := setupHeapFile(t)
	defer cleanup()

	rid := &heapfile.RecordID{PageID: 999, SlotID: 0}
	_, err := hf.GetRecord(rid)
	if err == nil {
		t.Errorf("GetRecord() expected error for non-existent page, got nil")
	}
}

func TestHeapFileGetRecordInvalidSlot(t *testing.T) {
	hf, _, _, cleanup := setupHeapFile(t)
	defer cleanup()

	rec := record.Record([]byte("data"))
	rid, err := hf.InsertRecord(rec)
	if err != nil {
		t.Fatalf("InsertRecord() error = %v", err)
	}

	invalidRid := &heapfile.RecordID{PageID: rid.PageID, SlotID: 999}
	_, err = hf.GetRecord(invalidRid)
	if err == nil {
		t.Errorf("GetRecord() expected error for invalid slot, got nil")
	}
}

func TestHeapFileInsertAfterGetRecord(t *testing.T) {
	hf, _, _, cleanup := setupHeapFile(t)
	defer cleanup()

	rec1 := record.Record([]byte("record1"))
	rid1, err := hf.InsertRecord(rec1)
	if err != nil {
		t.Fatalf("InsertRecord() error = %v", err)
	}

	rec2 := record.Record([]byte("record2"))
	rid2, err := hf.InsertRecord(rec2)
	if err != nil {
		t.Fatalf("InsertRecord() error = %v", err)
	}

	got1, _ := hf.GetRecord(rid1)
	got2, _ := hf.GetRecord(rid2)

	if string(got1) != "record1" {
		t.Errorf("GetRecord(rid1) = %q, want %q", string(got1), "record1")
	}

	if string(got2) != "record2" {
		t.Errorf("GetRecord(rid2) = %q, want %q", string(got2), "record2")
	}
}

func TestHeapFileInsertLargeRecord(t *testing.T) {
	hf, _, _, cleanup := setupHeapFile(t)
	defer cleanup()

	largeData := make([]byte, 4000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	rec := record.Record(largeData)

	rid, err := hf.InsertRecord(rec)
	if err != nil {
		t.Fatalf("InsertRecord() error = %v", err)
	}

	got, err := hf.GetRecord(rid)
	if err != nil {
		t.Fatalf("GetRecord() error = %v", err)
	}

	if len(got) != 4000 {
		t.Fatalf("len(record) = %d, want 4000", len(got))
	}

	for i := range got {
		if got[i] != byte(i%256) {
			t.Errorf("byte[%d] = %d, want %d", i, got[i], byte(i%256))
			break
		}
	}
}

func TestHeapFileMultiplePagesInsertAndGet(t *testing.T) {
	hf, _, metadata, cleanup := setupHeapFile(t)
	defer cleanup()

	for i := 0; i < 10; i++ {
		rec := record.Record([]byte("test data for insert"))
		_, err := hf.InsertRecord(rec)
		if err != nil {
			t.Fatalf("InsertRecord(%d) error = %v", i, err)
		}
	}

	if len(metadata.PageIDs) < 1 {
		t.Errorf("len(PageIDs) = %d, want at least 1", len(metadata.PageIDs))
	}

	for _, pageID := range metadata.PageIDs {
		rid := &heapfile.RecordID{PageID: pageID, SlotID: 0}
		_, err := hf.GetRecord(rid)
		if err != nil {
			t.Errorf("GetRecord(PageID=%d, SlotID=0) error = %v", pageID, err)
		}
	}
}

func TestHeapFileInsertZeroLengthRecord(t *testing.T) {
	hf, _, _, cleanup := setupHeapFile(t)
	defer cleanup()

	rec := record.Record{}
	_, err := hf.InsertRecord(rec)
	if err == nil {
		t.Errorf("InsertRecord() expected error for zero-length record, got nil")
	}
}
