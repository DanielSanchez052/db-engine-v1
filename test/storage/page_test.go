package storage_test

import (
	"bytes"
	"errors"
	"testing"

	"db-engine-v1/internal/storage"
	"db-engine-v1/internal/storage/page"
	"db-engine-v1/internal/storage/record"
	"db-engine-v1/internal/storage/slot"
)

func TestPageSerialize(t *testing.T) {
	p := page.NewPage(1, page.DataPage)

	bytes, err := p.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	if len(bytes) != storage.PageSize {
		t.Errorf("Serialize() len = %d, want %d", len(bytes), storage.PageSize)
	}
}

func TestPageDeserialize(t *testing.T) {
	p := page.NewPage(1, page.DataPage)

	bytes, err := p.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	p2, err := page.NewPageFromBytes(bytes)
	if err != nil {
		t.Fatalf("NewPageFromBytes() error = %v, want nil", err)
	}

	if *p.Header != *p2.Header {
		t.Errorf("Header mismatch: got %+v, want %+v", p2.Header, p.Header)
	}
}

func TestPagePayloadPersistence(t *testing.T) {
	p := page.NewPage(1, page.DataPage)

	p.Payload[0] = 10
	p.Payload[100] = 50

	bytes, err := p.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	p2, err := page.NewPageFromBytes(bytes)
	if err != nil {
		t.Fatalf("NewPageFromBytes() error = %v, want nil", err)
	}

	if p2.Payload[0] != 10 {
		t.Errorf("Payload[0] = %d, want 10", p2.Payload[0])
	}

	if p2.Payload[100] != 50 {
		t.Errorf("Payload[100] = %d, want 50", p2.Payload[100])
	}
}

func TestInvalidPageSize(t *testing.T) {
	data := make([]byte, 100)

	_, err := page.NewPageFromBytes(data)
	if err == nil {
		t.Errorf("NewPageFromBytes() expected error for invalid size, got nil")
	}

	if !errors.Is(err, page.ErrPageSizeMismatch) {
		t.Errorf("NewPageFromBytes() error = %v, want %v", err, page.ErrPageSizeMismatch)
	}
}

func TestAvailableSpace(t *testing.T) {
	p := page.NewPage(1, page.DataPage)

	// Initially: FreeSpaceOffset = PageSize (4096), SlotCount = 0
	// AvailableSpace = 4096 - (64 + 0*4) = 4032
	expected := uint16(storage.PageSize - storage.PageHeaderSize)
	if p.AvailableSpace() != expected {
		t.Errorf("AvailableSpace() = %d, want %d", p.AvailableSpace(), expected)
	}

	// Add some slots
	p.Header.SlotCount = 5
	p.Header.FreeSpaceOffset = 4000
	// AvailableSpace = 4000 - (64 + 5*4) = 4000 - 84 = 3916
	expected = uint16(4000 - (storage.PageHeaderSize + 5*storage.SlotSize))
	if p.AvailableSpace() != expected {
		t.Errorf("AvailableSpace() = %d, want %d", p.AvailableSpace(), expected)
	}
}

func TestCanFit(t *testing.T) {
	p := page.NewPage(1, page.DataPage)

	// Initially has plenty of space
	if !p.CanFit(100) {
		t.Errorf("CanFit(100) = false, want true")
	}

	// Record too large
	if p.CanFit(storage.PageSize) {
		t.Errorf("CanFit(PageSize) = true, want false")
	}

	// Set limited space
	p.Header.FreeSpaceOffset = 100
	p.Header.SlotCount = 0
	// AvailableSpace = 100 - 64 = 36
	// CanFit(30) should be true (30 + 4 = 34 <= 36)
	if !p.CanFit(30) {
		t.Errorf("CanFit(30) = false, want true")
	}

	// CanFit(35) should be false (35 + 4 = 39 > 36)
	if p.CanFit(35) {
		t.Errorf("CanFit(35) = true, want false")
	}
}

func TestSlotOffset(t *testing.T) {
	p := page.NewPage(1, page.DataPage)

	// Slot 0 should be at PageHeaderSize
	if p.SlotOffset(0) != storage.PageHeaderSize {
		t.Errorf("SlotOffset(0) = %d, want %d", p.SlotOffset(0), storage.PageHeaderSize)
	}

	// Slot 1 should be at PageHeaderSize + SlotSize
	if p.SlotOffset(1) != storage.PageHeaderSize+storage.SlotSize {
		t.Errorf("SlotOffset(1) = %d, want %d", p.SlotOffset(1), storage.PageHeaderSize+storage.SlotSize)
	}

	// Slot 5 should be at PageHeaderSize + 5*SlotSize
	expected := uint16(storage.PageHeaderSize + 5*storage.SlotSize)
	if p.SlotOffset(5) != expected {
		t.Errorf("SlotOffset(5) = %d, want %d", p.SlotOffset(5), expected)
	}
}

func TestWriteSlot(t *testing.T) {
	p := page.NewPage(1, page.DataPage)
	s := &slot.Slot{RecordOffset: 100, RecordLength: 50}

	// Write slot at ID 0 (should work when SlotCount is 0)
	err := p.WriteSlot(0, s)
	if err != nil {
		t.Fatalf("writeSlot(0) error = %v, want nil", err)
	}

	// Verify slot was written
	slotData := p.Payload[0:storage.SlotSize]
	got, err := slot.NewSlotFromBytes(slotData)
	if err != nil {
		t.Fatalf("NewSlotFromBytes error = %v", err)
	}
	if got.RecordOffset != s.RecordOffset || got.RecordLength != s.RecordLength {
		t.Errorf("written slot mismatch: got %+v, want %+v", got, s)
	}

	// Try to write at invalid slot ID (greater than SlotCount)
	p.Header.SlotCount = 1
	err = p.WriteSlot(2, s)
	if err == nil {
		t.Errorf("writeSlot(2) expected error, got nil")
	}
	if !errors.Is(err, page.ErrInvalidSlotID) {
		t.Errorf("writeSlot(2) error = %v, want %v", err, page.ErrInvalidSlotID)
	}
}

func TestReadSlot(t *testing.T) {
	p := page.NewPage(1, page.DataPage)
	s := &slot.Slot{RecordOffset: 200, RecordLength: 75}

	// Write a slot first
	p.Header.SlotCount = 1
	err := p.WriteSlot(0, s)
	if err != nil {
		t.Fatalf("writeSlot(0) error = %v", err)
	}

	// Read it back
	got, err := p.ReadSlot(0)
	if err != nil {
		t.Fatalf("readSlot(0) error = %v, want nil", err)
	}

	if got.RecordOffset != s.RecordOffset || got.RecordLength != s.RecordLength {
		t.Errorf("readSlot(0) mismatch: got %+v, want %+v", got, s)
	}

	// Try to read invalid slot ID
	_, err = p.ReadSlot(1)
	if err == nil {
		t.Errorf("readSlot(1) expected error, got nil")
	}
	if !errors.Is(err, page.ErrInvalidSlotID) {
		t.Errorf("readSlot(1) error = %v, want %v", err, page.ErrInvalidSlotID)
	}
}

func TestInsertRecord(t *testing.T) {
	p := page.NewPage(1, page.DataPage)
	record := record.Record("Hello, World!")

	slotID, err := p.InsertRecord(record)
	if err != nil {
		t.Fatalf("InsertRecord() error = %v, want nil", err)
	}

	if slotID != 0 {
		t.Errorf("InsertRecord() slotID = %d, want 0", slotID)
	}

	// Verify RecordCount and SlotCount were incremented
	if p.Header.RecordCount != 1 {
		t.Errorf("RecordCount = %d, want 1", p.Header.RecordCount)
	}
	if p.Header.SlotCount != 1 {
		t.Errorf("SlotCount = %d, want 1", p.Header.SlotCount)
	}

	// Verify the slot was written correctly
	slot, err := p.ReadSlot(0)
	if err != nil {
		t.Fatalf("ReadSlot() error = %v", err)
	}
	if slot.RecordLength != record.Size() {
		t.Errorf("slot.RecordLength = %d, want %d", slot.RecordLength, record.Size())
	}

	// Verify the record data was written
	payloadOffset := slot.RecordOffset - storage.PageHeaderSize
	writtenData := p.Payload[payloadOffset : payloadOffset+record.Size()]
	if !bytes.Equal(writtenData, record) {
		t.Errorf("record data mismatch: got %v, want %v", writtenData, record)
	}
}

func TestInsertMultipleRecords(t *testing.T) {
	p := page.NewPage(1, page.DataPage)

	record1 := record.Record("First")
	record2 := record.Record("Second")
	record3 := record.Record("Third")

	slotID1, err := p.InsertRecord(record1)
	if err != nil {
		t.Fatalf("InsertRecord(1) error = %v", err)
	}

	slotID2, err := p.InsertRecord(record2)
	if err != nil {
		t.Fatalf("InsertRecord(2) error = %v", err)
	}

	slotID3, err := p.InsertRecord(record3)
	if err != nil {
		t.Fatalf("InsertRecord(3) error = %v", err)
	}

	if slotID1 != 0 || slotID2 != 1 || slotID3 != 2 {
		t.Errorf("slotIDs = %d, %d, %d, want 0, 1, 2", slotID1, slotID2, slotID3)
	}

	if p.Header.RecordCount != 3 {
		t.Errorf("RecordCount = %d, want 3", p.Header.RecordCount)
	}

	// Verify each record can be read back
	for i, expectedRecord := range []record.Record{record1, record2, record3} {
		slot, err := p.ReadSlot(uint16(i))
		if err != nil {
			t.Fatalf("ReadSlot(%d) error = %v", i, err)
		}

		payloadOffset := slot.RecordOffset - storage.PageHeaderSize
		writtenData := p.Payload[payloadOffset : payloadOffset+slot.RecordLength]
		if !bytes.Equal(writtenData, expectedRecord) {
			t.Errorf("record %d data mismatch", i)
		}
	}
}

func TestInsertRecordUpdatesFreeSpaceOffset(t *testing.T) {
	p := page.NewPage(1, page.DataPage)
	initialOffset := p.Header.FreeSpaceOffset

	record := record.Record("Test data")
	_, err := p.InsertRecord(record)
	if err != nil {
		t.Fatalf("InsertRecord() error = %v", err)
	}

	expectedOffset := initialOffset - record.Size()
	if p.Header.FreeSpaceOffset != expectedOffset {
		t.Errorf("FreeSpaceOffset = %d, want %d", p.Header.FreeSpaceOffset, expectedOffset)
	}
}

func TestInsertRecordEmpty(t *testing.T) {
	p := page.NewPage(1, page.DataPage)
	record := record.Record{}

	_, err := p.InsertRecord(record)
	if err == nil {
		t.Errorf("InsertRecord() expected error for empty record, got nil")
	}
	if !errors.Is(err, page.ErrEmptyRecord) {
		t.Errorf("InsertRecord() error = %v, want %v", err, page.ErrEmptyRecord)
	}
}

func TestInsertRecordNotEnoughSpace(t *testing.T) {
	p := page.NewPage(1, page.DataPage)

	// Create a record that's too large
	largeRecord := make(record.Record, storage.PageSize)
	_, err := p.InsertRecord(largeRecord)
	if err == nil {
		t.Errorf("InsertRecord() expected error for large record, got nil")
	}
	if !errors.Is(err, page.ErrNotEnoughSpace) {
		t.Errorf("InsertRecord() error = %v, want %v", err, page.ErrNotEnoughSpace)
	}
}

func TestInsertRecordPageFull(t *testing.T) {
	p := page.NewPage(1, page.DataPage)

	// Fill the page with small records
	recordSize := uint16(10)
	maxRecords := (storage.PageSize - storage.PageHeaderSize) / (recordSize + storage.SlotSize)

	for i := 0; i < int(maxRecords); i++ {
		record := make(record.Record, recordSize)
		_, err := p.InsertRecord(record)
		if err != nil {
			t.Fatalf("InsertRecord(%d) error = %v", i, err)
		}
	}

	// Try to insert one more record
	record := make(record.Record, recordSize)
	_, err := p.InsertRecord(record)
	if err == nil {
		t.Errorf("InsertRecord() expected error when page is full, got nil")
	}
	if !errors.Is(err, page.ErrNotEnoughSpace) {
		t.Errorf("InsertRecord() error = %v, want %v", err, page.ErrNotEnoughSpace)
	}
}

func TestInsertAndGetRecord(t *testing.T) {
	p := page.NewPage(1, page.DataPage)
	record := record.Record("Hello, World!")

	slotID, err := p.InsertRecord(record)
	if err != nil {
		t.Fatalf("InsertRecord() error = %v", err)
	}

	retrieved, err := p.GetRecord(slotID)
	if err != nil {
		t.Fatalf("GetRecord() error = %v", err)
	}

	if !bytes.Equal(retrieved, record) {
		t.Errorf("GetRecord() = %v, want %v", retrieved, record)
	}
}

func TestGetRecordMultiple(t *testing.T) {
	p := page.NewPage(1, page.DataPage)

	records := []record.Record{
		record.Record("First"),
		record.Record("Second"),
		record.Record("Third"),
	}

	slotIDs := make([]uint16, len(records))
	for i, rec := range records {
		slotID, err := p.InsertRecord(rec)
		if err != nil {
			t.Fatalf("InsertRecord(%d) error = %v", i, err)
		}
		slotIDs[i] = slotID
	}

	// Retrieve each record
	for i, expected := range records {
		retrieved, err := p.GetRecord(slotIDs[i])
		if err != nil {
			t.Fatalf("GetRecord(%d) error = %v", i, err)
		}

		if !bytes.Equal(retrieved, expected) {
			t.Errorf("GetRecord(%d) = %v, want %v", i, retrieved, expected)
		}
	}
}

func TestGetRecordInvalidSlot(t *testing.T) {
	p := page.NewPage(1, page.DataPage)

	_, err := p.GetRecord(0)
	if err == nil {
		t.Errorf("GetRecord(0) expected error for invalid slot, got nil")
	}
	if !errors.Is(err, page.ErrInvalidSlotID) {
		t.Errorf("GetRecord(0) error = %v, want %v", err, page.ErrInvalidSlotID)
	}

	_, err = p.GetRecord(999)
	if err == nil {
		t.Errorf("GetRecord(999) expected error for invalid slot, got nil")
	}
	if !errors.Is(err, page.ErrInvalidSlotID) {
		t.Errorf("GetRecord(999) error = %v, want %v", err, page.ErrInvalidSlotID)
	}
}

func TestGetRecordDeleted(t *testing.T) {
	p := page.NewPage(1, page.DataPage)
	record := record.Record("Test data")

	slotID, err := p.InsertRecord(record)
	if err != nil {
		t.Fatalf("InsertRecord() error = %v", err)
	}

	// Manually mark the slot as deleted
	slot, err := p.ReadSlot(slotID)
	if err != nil {
		t.Fatalf("ReadSlot() error = %v", err)
	}
	slot.RecordOffset = 0
	slot.RecordLength = 0
	err = p.WriteSlot(slotID, slot)
	if err != nil {
		t.Fatalf("WriteSlot() error = %v", err)
	}

	// Try to get the deleted record
	_, err = p.GetRecord(slotID)
	if err == nil {
		t.Errorf("GetRecord() expected error for deleted record, got nil")
	}
	if !errors.Is(err, page.ErrRecordDeleted) {
		t.Errorf("GetRecord() error = %v, want %v", err, page.ErrRecordDeleted)
	}
}

func TestDeleteRecord(t *testing.T) {
	p := page.NewPage(1, page.DataPage)
	record := record.Record("Test data")

	slotID, err := p.InsertRecord(record)
	if err != nil {
		t.Fatalf("InsertRecord() error = %v", err)
	}

	initialRecordCount := p.Header.RecordCount

	err = p.DeleteRecord(slotID)
	if err != nil {
		t.Fatalf("DeleteRecord() error = %v, want nil", err)
	}

	// Verify RecordCount was decremented
	if p.Header.RecordCount != initialRecordCount-1 {
		t.Errorf("RecordCount = %d, want %d", p.Header.RecordCount, initialRecordCount-1)
	}

	// Verify the slot is marked as deleted
	slot, err := p.ReadSlot(slotID)
	if err != nil {
		t.Fatalf("ReadSlot() error = %v", err)
	}
	if !slot.IsDeleted() {
		t.Errorf("slot.IsDeleted() = false, want true")
	}
}

func TestDeleteRecordInvalidSlot(t *testing.T) {
	p := page.NewPage(1, page.DataPage)

	err := p.DeleteRecord(0)
	if err == nil {
		t.Errorf("DeleteRecord(0) expected error for invalid slot, got nil")
	}
	if !errors.Is(err, page.ErrInvalidSlotID) {
		t.Errorf("DeleteRecord(0) error = %v, want %v", err, page.ErrInvalidSlotID)
	}

	err = p.DeleteRecord(999)
	if err == nil {
		t.Errorf("DeleteRecord(999) expected error for invalid slot, got nil")
	}
	if !errors.Is(err, page.ErrInvalidSlotID) {
		t.Errorf("DeleteRecord(999) error = %v, want %v", err, page.ErrInvalidSlotID)
	}
}

func TestDeleteTwice(t *testing.T) {
	p := page.NewPage(1, page.DataPage)
	record := record.Record("Test data")

	slotID, err := p.InsertRecord(record)
	if err != nil {
		t.Fatalf("InsertRecord() error = %v", err)
	}

	// First delete
	err = p.DeleteRecord(slotID)
	if err != nil {
		t.Fatalf("DeleteRecord() first call error = %v", err)
	}

	// Second delete should fail
	err = p.DeleteRecord(slotID)
	if err == nil {
		t.Errorf("DeleteRecord() second call expected error, got nil")
	}
	if !errors.Is(err, page.ErrRecordDeleted) {
		t.Errorf("DeleteRecord() second call error = %v, want %v", err, page.ErrRecordDeleted)
	}
}
