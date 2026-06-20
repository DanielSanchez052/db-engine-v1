package catalog_test

import (
	"testing"

	"db-engine-v1/internal/catalog"
)

func TestNewHeapMetadata(t *testing.T) {
	h := &catalog.HeapMetadata{
		Name:    "test_heap",
		PageIDs: nil,
	}

	if h.Name != "test_heap" {
		t.Errorf("Name = %q, want %q", h.Name, "test_heap")
	}

	if len(h.PageIDs) != 0 {
		t.Errorf("len(PageIDs) = %d, want 0", len(h.PageIDs))
	}
}

func TestHeapMetadataSize(t *testing.T) {
	t.Run("empty name", func(t *testing.T) {
		h := &catalog.HeapMetadata{Name: ""}
		expected := 2 + 0 + 4 + 0
		if h.Size() != expected {
			t.Errorf("Size() = %d, want %d", h.Size(), expected)
		}
	})

	t.Run("with name", func(t *testing.T) {
		h := &catalog.HeapMetadata{Name: "users"}
		expected := 2 + 5 + 4 + 0
		if h.Size() != expected {
			t.Errorf("Size() = %d, want %d", h.Size(), expected)
		}
	})

	t.Run("with name and pages", func(t *testing.T) {
		h := &catalog.HeapMetadata{Name: "users", PageIDs: []uint64{1, 2, 3}}
		expected := 2 + 5 + 4 + (3 * 8)
		if h.Size() != expected {
			t.Errorf("Size() = %d, want %d", h.Size(), expected)
		}
	})
}

func TestHeapMetadataSerializeAndDeserialize(t *testing.T) {
	original := &catalog.HeapMetadata{
		Name:    "test_heap",
		PageIDs: []uint64{1, 2, 3},
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	got, err := catalog.NewHeapMetadataFromBytes(data)
	if err != nil {
		t.Fatalf("NewHeapMetadataFromBytes() error = %v, want nil", err)
	}

	if got.Name != original.Name {
		t.Errorf("Name = %q, want %q", got.Name, original.Name)
	}

	if len(got.PageIDs) != len(original.PageIDs) {
		t.Fatalf("len(PageIDs) = %d, want %d", len(got.PageIDs), len(original.PageIDs))
	}

	for i, id := range original.PageIDs {
		if got.PageIDs[i] != id {
			t.Errorf("PageIDs[%d] = %d, want %d", i, got.PageIDs[i], id)
		}
	}
}

func TestHeapMetadataSerializeEmptyName(t *testing.T) {
	original := &catalog.HeapMetadata{
		Name:    "",
		PageIDs: []uint64{1, 2, 3},
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	got, err := catalog.NewHeapMetadataFromBytes(data)
	if err != nil {
		t.Fatalf("NewHeapMetadataFromBytes() error = %v, want nil", err)
	}

	if got.Name != "" {
		t.Errorf("Name = %q, want empty", got.Name)
	}

	if len(got.PageIDs) != len(original.PageIDs) {
		t.Fatalf("len(PageIDs) = %d, want %d", len(got.PageIDs), len(original.PageIDs))
	}
}

func TestHeapMetadataSerializeNoPages(t *testing.T) {
	original := &catalog.HeapMetadata{
		Name:    "empty_heap",
		PageIDs: []uint64{},
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	got, err := catalog.NewHeapMetadataFromBytes(data)
	if err != nil {
		t.Fatalf("NewHeapMetadataFromBytes() error = %v, want nil", err)
	}

	if got.Name != original.Name {
		t.Errorf("Name = %q, want %q", got.Name, original.Name)
	}

	if len(got.PageIDs) != 0 {
		t.Errorf("len(PageIDs) = %d, want 0", len(got.PageIDs))
	}
}

func TestHeapMetadataRoundTrip(t *testing.T) {
	original := &catalog.HeapMetadata{
		Name:    "round_trip_heap",
		PageIDs: []uint64{10, 20, 30, 40, 50},
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	got, err := catalog.NewHeapMetadataFromBytes(data)
	if err != nil {
		t.Fatalf("NewHeapMetadataFromBytes() error = %v, want nil", err)
	}

	if got.Name != original.Name {
		t.Errorf("Name = %q, want %q", got.Name, original.Name)
	}

	if len(got.PageIDs) != len(original.PageIDs) {
		t.Fatalf("len(PageIDs) = %d, want %d", len(got.PageIDs), len(original.PageIDs))
	}

	for i, id := range original.PageIDs {
		if got.PageIDs[i] != id {
			t.Errorf("PageIDs[%d] = %d, want %d", i, got.PageIDs[i], id)
		}
	}
}

func TestHeapMetadataAddPage(t *testing.T) {
	h := &catalog.HeapMetadata{
		Name:    "growing_heap",
		PageIDs: []uint64{1, 2},
	}

	h.AddPage(3)

	if len(h.PageIDs) != 3 {
		t.Fatalf("len(PageIDs) after AddPage = %d, want 3", len(h.PageIDs))
	}

	if h.PageIDs[2] != 3 {
		t.Errorf("PageIDs[2] = %d, want 3", h.PageIDs[2])
	}

	h.AddPage(4)
	if len(h.PageIDs) != 4 {
		t.Errorf("len(PageIDs) after second AddPage = %d, want 4", len(h.PageIDs))
	}
}

func TestHeapMetadataNameTooLong(t *testing.T) {
	longName := ""
	for i := 0; i < 256; i++ {
		longName += "a"
	}

	h := &catalog.HeapMetadata{Name: longName}

	_, err := h.Serialize()
	if err == nil {
		t.Errorf("Serialize() expected error for name length > 255, got nil")
	}
}

func TestNewHeapMetadataFromBytesInvalidData(t *testing.T) {
	t.Run("empty data", func(t *testing.T) {
		_, err := catalog.NewHeapMetadataFromBytes([]byte{})
		if err == nil {
			t.Errorf("NewHeapMetadataFromBytes() expected error for empty data, got nil")
		}
	})

	t.Run("truncated data", func(t *testing.T) {
		data := make([]byte, 4)
		_, err := catalog.NewHeapMetadataFromBytes(data)
		if err == nil {
			t.Errorf("NewHeapMetadataFromBytes() expected error for truncated data, got nil")
		}
	})
}

func TestHeapMetadataLargePageSet(t *testing.T) {
	pageIDs := make([]uint64, 100)
	for i := 0; i < 100; i++ {
		pageIDs[i] = uint64(i + 1)
	}

	original := &catalog.HeapMetadata{
		Name:    "large_heap",
		PageIDs: pageIDs,
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	got, err := catalog.NewHeapMetadataFromBytes(data)
	if err != nil {
		t.Fatalf("NewHeapMetadataFromBytes() error = %v, want nil", err)
	}

	if len(got.PageIDs) != 100 {
		t.Fatalf("len(PageIDs) = %d, want 100", len(got.PageIDs))
	}

	for i := 0; i < 100; i++ {
		if got.PageIDs[i] != pageIDs[i] {
			t.Errorf("PageIDs[%d] = %d, want %d", i, got.PageIDs[i], pageIDs[i])
		}
	}
}
