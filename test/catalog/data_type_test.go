package catalog_test

import (
	"testing"

	"db-engine-v1/internal/catalog"
)

func TestDataTypeValues(t *testing.T) {
	if catalog.TypeInt32Type != 0 {
		t.Errorf("TypeInt32Type = %d, want 0", catalog.TypeInt32Type)
	}

	if catalog.TypeStringType != 1 {
		t.Errorf("TypeStringType = %d, want 1", catalog.TypeStringType)
	}

	if catalog.TypeBoolType != 2 {
		t.Errorf("TypeBoolType = %d, want 2", catalog.TypeBoolType)
	}
}

func TestDataTypeIsValid(t *testing.T) {
	t.Run("TypeInt32Type", func(t *testing.T) {
		if !catalog.TypeInt32Type.IsValid() {
			t.Errorf("TypeInt32Type.IsValid() = false, want true")
		}
	})

	t.Run("TypeStringType", func(t *testing.T) {
		if !catalog.TypeStringType.IsValid() {
			t.Errorf("TypeStringType.IsValid() = false, want true")
		}
	})

	t.Run("TypeBoolType", func(t *testing.T) {
		if !catalog.TypeBoolType.IsValid() {
			t.Errorf("TypeBoolType.IsValid() = false, want true")
		}
	})
}

func TestDataTypeIsValidOutOfRange(t *testing.T) {
	t.Run("below range", func(t *testing.T) {
		var dt catalog.DataType = 255
		if dt.IsValid() {
			t.Errorf("DataType(255).IsValid() = true, want false")
		}
	})

	t.Run("above range", func(t *testing.T) {
		var dt catalog.DataType = 3
		if dt.IsValid() {
			t.Errorf("DataType(3).IsValid() = true, want false")
		}
	})
}
