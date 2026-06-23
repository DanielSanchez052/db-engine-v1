package storage_test

import (
	"errors"
	"testing"

	"db-engine-v1/internal/storage/heapfile"
	"db-engine-v1/internal/storage/record"
)

func TestHeapFileIteratorNew(t *testing.T) {
	hf, _, _, cleanup := setupHeapFile(t)
	defer cleanup()

	it := heapfile.NewIterator(hf)
	if it == nil {
		t.Fatalf("NewIterator() returned nil")
	}
}

func TestHeapFileIteratorEmptyHeap(t *testing.T) {
	hf, _, _, cleanup := setupHeapFile(t)
	defer cleanup()

	it := heapfile.NewIterator(hf)

	_, _, err := it.Next()
	if !errors.Is(err, heapfile.ErrIteratorDone) {
		t.Errorf("Next() error = %v, want %v", err, heapfile.ErrIteratorDone)
	}
}

func TestHeapFileIteratorSingleRecord(t *testing.T) {
	hf, _, _, cleanup := setupHeapFile(t)
	defer cleanup()

	rec := record.Record([]byte("hello"))
	rid, _, err := hf.InsertRecord(rec)
	if err != nil {
		t.Fatalf("InsertRecord() error = %v", err)
	}

	it := heapfile.NewIterator(hf)

	gotRec, gotRID, err := it.Next()
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}

	if string(gotRec) != "hello" {
		t.Errorf("record = %q, want %q", string(gotRec), "hello")
	}

	if gotRID.PageID != rid.PageID || gotRID.SlotID != rid.SlotID {
		t.Errorf("RecordID = %+v, want %+v", gotRID, rid)
	}

	// Should be done
	_, _, err = it.Next()
	if !errors.Is(err, heapfile.ErrIteratorDone) {
		t.Errorf("Next() after last record error = %v, want %v", err, heapfile.ErrIteratorDone)
	}
}

func TestHeapFileIteratorMultipleRecords(t *testing.T) {
	hf, _, _, cleanup := setupHeapFile(t)
	defer cleanup()

	records := []string{"Alice", "Bob", "Charlie"}
	var inserted []*heapfile.RecordID

	for _, name := range records {
		rid, _, err := hf.InsertRecord(record.Record([]byte(name)))
		if err != nil {
			t.Fatalf("InsertRecord(%q) error = %v", name, err)
		}
		inserted = append(inserted, rid)
	}

	it := heapfile.NewIterator(hf)

	for i, expected := range records {
		gotRec, gotRID, err := it.Next()
		if err != nil {
			t.Fatalf("Next(%d) error = %v, want nil", i, err)
		}

		if string(gotRec) != expected {
			t.Errorf("record[%d] = %q, want %q", i, string(gotRec), expected)
		}

		if gotRID.PageID != inserted[i].PageID || gotRID.SlotID != inserted[i].SlotID {
			t.Errorf("RecordID[%d] = %+v, want %+v", i, gotRID, inserted[i])
		}
	}

	// Should be done
	_, _, err := it.Next()
	if !errors.Is(err, heapfile.ErrIteratorDone) {
		t.Errorf("Next() after all records error = %v, want %v", err, heapfile.ErrIteratorDone)
	}
}

func TestHeapFileIteratorMultiplePages(t *testing.T) {
	// Disable parallel execution to avoid test pollution
	hf, _, metadata, cleanup := setupHeapFile(t)
	defer cleanup()

	// Insert large records to force multiple pages
	largeData := make([]byte, 3000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	expectedCount := 3
	for i := 0; i < expectedCount; i++ {
		_, _, err := hf.InsertRecord(record.Record(largeData))
		if err != nil {
			t.Fatalf("InsertRecord(%d) error = %v", i, err)
		}
	}

	if len(metadata.PageIDs) < expectedCount {
		t.Fatalf("len(PageIDs) = %d, want at least %d (large records should create multiple pages)", len(metadata.PageIDs), expectedCount)
	}

	it := heapfile.NewIterator(hf)

	count := 0
	for {
		gotRec, _, err := it.Next()
		if errors.Is(err, heapfile.ErrIteratorDone) {
			break
		}
		if err != nil {
			t.Fatalf("Next(%d) error = %v", count, err)
		}
		if len(gotRec) != 3000 {
			t.Errorf("record[%d] len = %d, want 3000", count, len(gotRec))
		}
		count++
	}

	if count != expectedCount {
		t.Errorf("iterated %d records, want %d", count, expectedCount)
	}
}

func TestHeapFileIteratorSkipsDeletedRecords(t *testing.T) {
	hf, pg, _, cleanup := setupHeapFile(t)
	defer cleanup()

	rid1, _, _ := hf.InsertRecord(record.Record([]byte("delete me")))
	hf.InsertRecord(record.Record([]byte("keep me")))
	rid3, _, _ := hf.InsertRecord(record.Record([]byte("delete me too")))

	// Delete first and third records
	page1, err := hf.GetPage(rid1.PageID)
	if err != nil {
		t.Fatalf("GetPage() error = %v", err)
	}

	err = page1.DeleteRecord(rid1.SlotID)
	if err != nil {
		t.Fatalf("DeleteRecord(slot=%d) error = %v", rid1.SlotID, err)
	}

	err = page1.DeleteRecord(rid3.SlotID)
	if err != nil {
		t.Fatalf("DeleteRecord(slot=%d) error = %v", rid3.SlotID, err)
	}

	// Persist the modified page back to the pager
	err = pg.SavePage(page1)
	if err != nil {
		t.Fatalf("SavePage() error = %v", err)
	}

	it := heapfile.NewIterator(hf)

	// Should only get the second record
	gotRec, gotRID, err := it.Next()
	if err != nil {
		t.Fatalf("Next() error = %v, want nil", err)
	}

	if string(gotRec) != "keep me" {
		t.Errorf("record = %q, want %q", string(gotRec), "keep me")
	}

	if gotRID.SlotID != 1 {
		t.Errorf("SlotID = %d, want 1 (second slot)", gotRID.SlotID)
	}

	// Should be done
	_, _, err = it.Next()
	if !errors.Is(err, heapfile.ErrIteratorDone) {
		t.Errorf("Next() after all records error = %v, want %v", err, heapfile.ErrIteratorDone)
	}
}

func TestHeapFileIteratorRecordsInOrder(t *testing.T) {
	hf, _, _, cleanup := setupHeapFile(t)
	defer cleanup()

	for i := 0; i < 5; i++ {
		data := []byte{byte('a' + i)}
		_, _, err := hf.InsertRecord(record.Record(data))
		if err != nil {
			t.Fatalf("InsertRecord(%d) error = %v", i, err)
		}
	}

	it := heapfile.NewIterator(hf)

	for i := 0; i < 5; i++ {
		gotRec, _, err := it.Next()
		if err != nil {
			t.Fatalf("Next(%d) error = %v", i, err)
		}
		if len(gotRec) != 1 || gotRec[0] != byte('a'+i) {
			t.Errorf("record[%d] = %v, want {%d}", i, gotRec, 'a'+i)
		}
	}
}

func TestHeapFileIteratorCallAfterDone(t *testing.T) {
	hf, _, _, cleanup := setupHeapFile(t)
	defer cleanup()

	it := heapfile.NewIterator(hf)

	// Empty heap, first call returns done
	_, _, err := it.Next()
	if !errors.Is(err, heapfile.ErrIteratorDone) {
		t.Fatalf("Next() error = %v, want %v", err, heapfile.ErrIteratorDone)
	}

	// Subsequent calls should also return done
	for i := 0; i < 3; i++ {
		_, _, err = it.Next()
		if !errors.Is(err, heapfile.ErrIteratorDone) {
			t.Errorf("Next() after done (call %d) error = %v, want %v", i, err, heapfile.ErrIteratorDone)
		}
	}
}

func TestHeapFileIteratorAllDeletedReturnsDone(t *testing.T) {
	hf, pg, _, cleanup := setupHeapFile(t)
	defer cleanup()

	rid, _, _ := hf.InsertRecord(record.Record([]byte("lonely")))

	page1, err := hf.GetPage(rid.PageID)
	if err != nil {
		t.Fatalf("GetPage() error = %v", err)
	}

	err = page1.DeleteRecord(rid.SlotID)
	if err != nil {
		t.Fatalf("DeleteRecord() error = %v", err)
	}

	err = pg.SavePage(page1)
	if err != nil {
		t.Fatalf("SavePage() error = %v", err)
	}

	it := heapfile.NewIterator(hf)
	_, _, err = it.Next()
	if !errors.Is(err, heapfile.ErrIteratorDone) {
		t.Errorf("Next() with all deleted error = %v, want %v", err, heapfile.ErrIteratorDone)
	}
}
