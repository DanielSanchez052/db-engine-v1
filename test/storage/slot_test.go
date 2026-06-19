package storage_test

import (
	"errors"
	"testing"

	"db-engine-v1/internal/storage/slot"
)

func TestSlotSerialization(t *testing.T) {
	original := &slot.Slot{
		RecordOffset: 100,
		RecordLength: 50,
	}

	bytes, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	got, err := slot.NewSlotFromBytes(bytes)
	if err != nil {
		t.Fatalf("NewSlotFromBytes() error = %v, want nil", err)
	}

	if *got != *original {
		t.Errorf("round trip mismatch: got %+v, want %+v", got, original)
	}
}

func TestDeletedSlot(t *testing.T) {
	s := &slot.Slot{}

	if !s.IsDeleted() {
		t.Errorf("IsDeleted() = false, want true for empty slot")
	}
}

func TestValidateOffsetOutOfRange(t *testing.T) {
	s := &slot.Slot{
		RecordOffset: 5000,
		RecordLength: 10,
	}

	err := s.Validate()
	if err == nil {
		t.Errorf("Validate() expected error for RecordOffset = 5000, got nil")
	}

	if !errors.Is(err, slot.ErrInvalidRecordOffset) {
		t.Errorf("Validate() error = %v, want %v", err, slot.ErrInvalidRecordOffset)
	}
}

func TestValidateRecordOverflow(t *testing.T) {
	s := &slot.Slot{
		RecordOffset: 4090,
		RecordLength: 20,
	}

	err := s.Validate()
	if err == nil {
		t.Errorf("Validate() expected error for RecordOffset + RecordLength > PageSize, got nil")
	}

	if !errors.Is(err, slot.ErrRecordExceedsPage) {
		t.Errorf("Validate() error = %v, want %v", err, slot.ErrRecordExceedsPage)
	}
}
