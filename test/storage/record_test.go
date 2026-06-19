package storage_test

import (
	"testing"

	"db-engine-v1/internal/storage/record"
)

func TestSize(t *testing.T) {
	r := record.Record([]byte("Daniel"))

	if r.Size() != 6 {
		t.Fatalf("Size() = %d, want 6", r.Size())
	}
}

func TestEmptyRecord(t *testing.T) {
	r := record.Record{}

	if r.Size() != 0 {
		t.Fatalf("Size() = %d, want 0", r.Size())
	}
}
