package tuple_test

import (
	"testing"

	"db-engine-v1/internal/catalog"
	"db-engine-v1/internal/storage/tuple"
)

func TestNewBoolValue(t *testing.T) {
	v := tuple.NewBoolValue(true)
	if v.Value != true {
		t.Errorf("Value = %v, want true", v.Value)
	}

	v2 := tuple.NewBoolValue(false)
	if v2.Value != false {
		t.Errorf("Value = %v, want false", v2.Value)
	}
}

func TestBoolValueType(t *testing.T) {
	v := tuple.NewBoolValue(true)
	if v.Type() != catalog.TypeBoolType {
		t.Errorf("Type() = %v, want %v", v.Type(), catalog.TypeBoolType)
	}
}

func TestBoolValueSize(t *testing.T) {
	v := tuple.NewBoolValue(true)
	if v.Size() != catalog.BoolSize {
		t.Errorf("Size() = %d, want %d", v.Size(), catalog.BoolSize)
	}

	v2 := tuple.NewBoolValue(false)
	if v2.Size() != catalog.BoolSize {
		t.Errorf("Size() = %d, want %d", v2.Size(), catalog.BoolSize)
	}
}

func TestBoolValueSerialize(t *testing.T) {
	v := tuple.NewBoolValue(true)
	data, err := v.Serialize()
	if err != nil {
		t.Fatalf("Serialize(true) error = %v", err)
	}
	if len(data) != 1 {
		t.Fatalf("len = %d, want 1", len(data))
	}
	if data[0] != 1 {
		t.Errorf("Serialize(true) = %d, want 1", data[0])
	}

	v2 := tuple.NewBoolValue(false)
	data2, err := v2.Serialize()
	if err != nil {
		t.Fatalf("Serialize(false) error = %v", err)
	}
	if len(data2) != 1 {
		t.Fatalf("len = %d, want 1", len(data2))
	}
	if data2[0] != 0 {
		t.Errorf("Serialize(false) = %d, want 0", data2[0])
	}
}

func TestBoolValueFromBytesTrue(t *testing.T) {
	v, err := tuple.NewBoolValueFromBytes([]byte{1})
	if err != nil {
		t.Fatalf("NewBoolValueFromBytes([1]) error = %v", err)
	}
	if v.Value != true {
		t.Errorf("Value = %v, want true", v.Value)
	}
}

func TestBoolValueFromBytesFalse(t *testing.T) {
	v, err := tuple.NewBoolValueFromBytes([]byte{0})
	if err != nil {
		t.Fatalf("NewBoolValueFromBytes([0]) error = %v", err)
	}
	if v.Value != false {
		t.Errorf("Value = %v, want false", v.Value)
	}
}

func TestBoolValueFromBytesInvalid(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		v, err := tuple.NewBoolValueFromBytes(nil)
		if v != nil {
			t.Errorf("NewBoolValueFromBytes(nil) = %v, want nil", v)
		}
		if err != tuple.ErrInvalidValue {
			t.Errorf("error = %v, want %v", err, tuple.ErrInvalidValue)
		}
	})

	t.Run("empty", func(t *testing.T) {
		v, err := tuple.NewBoolValueFromBytes([]byte{})
		if v != nil {
			t.Errorf("NewBoolValueFromBytes(empty) = %v, want nil", v)
		}
		if err != tuple.ErrInvalidValue {
			t.Errorf("error = %v, want %v", err, tuple.ErrInvalidValue)
		}
	})

	t.Run("too_large", func(t *testing.T) {
		v, err := tuple.NewBoolValueFromBytes([]byte{0, 0})
		if v != nil {
			t.Errorf("NewBoolValueFromBytes(2bytes) = %v, want nil", v)
		}
		if err != tuple.ErrInvalidValue {
			t.Errorf("error = %v, want %v", err, tuple.ErrInvalidValue)
		}
	})

	t.Run("invalid_value_2", func(t *testing.T) {
		v, err := tuple.NewBoolValueFromBytes([]byte{2})
		if v != nil {
			t.Errorf("NewBoolValueFromBytes([2]) = %v, want nil", v)
		}
		if err != tuple.ErrInvalidValue {
			t.Errorf("error = %v, want %v", err, tuple.ErrInvalidValue)
		}
	})

	t.Run("invalid_value_255", func(t *testing.T) {
		v, err := tuple.NewBoolValueFromBytes([]byte{255})
		if v != nil {
			t.Errorf("NewBoolValueFromBytes([255]) = %v, want nil", v)
		}
		if err != tuple.ErrInvalidValue {
			t.Errorf("error = %v, want %v", err, tuple.ErrInvalidValue)
		}
	})
}

func TestBoolValueRoundTrip(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		original := tuple.NewBoolValue(true)
		data, _ := original.Serialize()
		restored, err := tuple.NewBoolValueFromBytes(data)
		if err != nil {
			t.Fatalf("NewBoolValueFromBytes() error = %v", err)
		}
		if restored.Value != true {
			t.Errorf("round trip = %v, want true", restored.Value)
		}
	})

	t.Run("false", func(t *testing.T) {
		original := tuple.NewBoolValue(false)
		data, _ := original.Serialize()
		restored, err := tuple.NewBoolValueFromBytes(data)
		if err != nil {
			t.Fatalf("NewBoolValueFromBytes() error = %v", err)
		}
		if restored.Value != false {
			t.Errorf("round trip = %v, want false", restored.Value)
		}
	})
}

func TestBoolValueString(t *testing.T) {
	v1 := tuple.NewBoolValue(true)
	if v1.String() != "true" {
		t.Errorf("String(true) = %q, want %q", v1.String(), "true")
	}

	v2 := tuple.NewBoolValue(false)
	if v2.String() != "false" {
		t.Errorf("String(false) = %q, want %q", v2.String(), "false")
	}
}

func TestBoolValueImplementsValue(t *testing.T) {
	var v tuple.Value = tuple.NewBoolValue(true)

	if v.Type() != catalog.TypeBoolType {
		t.Errorf("Type() = %v, want %v", v.Type(), catalog.TypeBoolType)
	}

	data, err := v.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	if len(data) != 1 {
		t.Errorf("Serialize() len = %d, want 1", len(data))
	}

	if v.Size() != catalog.BoolSize {
		t.Errorf("Size() = %d, want %d", v.Size(), catalog.BoolSize)
	}
}

func TestBoolValueZeroValue(t *testing.T) {
	v := tuple.BoolValue{}

	if v.Value != false {
		t.Errorf("default Value = %v, want false", v.Value)
	}

	if v.Type() != catalog.TypeBoolType {
		t.Errorf("Type() = %v, want %v", v.Type(), catalog.TypeBoolType)
	}

	if v.Size() != catalog.BoolSize {
		t.Errorf("Size() = %d, want %d", v.Size(), catalog.BoolSize)
	}

	data, err := v.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	if data[0] != 0 {
		t.Errorf("zero value serialized = %d, want 0", data[0])
	}

	restored, err := tuple.NewBoolValueFromBytes(data)
	if err != nil {
		t.Fatalf("NewBoolValueFromBytes() error = %v", err)
	}
	if restored.Value != false {
		t.Errorf("zero value round trip = %v, want false", restored.Value)
	}
}
