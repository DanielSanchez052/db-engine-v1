package tuple_test

import (
	"encoding/binary"
	"math"
	"strconv"
	"testing"

	"db-engine-v1/internal/catalog"
	"db-engine-v1/internal/storage/tuple"
)

func TestNewInt32Value(t *testing.T) {
	v := tuple.Int32Value{Value: 42}
	if v.Value != 42 {
		t.Errorf("Value = %d, want %d", v.Value, 42)
	}
}

func TestInt32ValueType(t *testing.T) {
	v := tuple.Int32Value{Value: 100}
	if v.Type() != catalog.TypeInt32Type {
		t.Errorf("Type() = %v, want %v", v.Type(), catalog.TypeInt32Type)
	}
}

func TestInt32ValueSize(t *testing.T) {
	v := tuple.Int32Value{Value: 0}
	if v.Size() != catalog.Uint32Size {
		t.Errorf("Size() = %d, want %d", v.Size(), catalog.Uint32Size)
	}
}

func TestInt32ValueSerialize(t *testing.T) {
	tests := []struct {
		name  string
		value int32
	}{
		{"zero", 0},
		{"positive", 42},
		{"negative", -1},
		{"max", math.MaxInt32},
		{"min", math.MinInt32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tuple.Int32Value{Value: tt.value}
			data, err := v.Serialize()
			if err != nil {
				t.Fatalf("Serialize() error = %v, want nil", err)
			}

			if len(data) != catalog.Uint32Size {
				t.Fatalf("len(data) = %d, want %d", len(data), catalog.Uint32Size)
			}

			got := int32(binary.LittleEndian.Uint32(data))
			if got != tt.value {
				t.Errorf("serialized value = %d, want %d", got, tt.value)
			}
		})
	}
}

func TestInt32ValueFromBytes(t *testing.T) {
	tests := []struct {
		name  string
		value int32
	}{
		{"zero", 0},
		{"positive", 42},
		{"negative", -1},
		{"max", math.MaxInt32},
		{"min", math.MinInt32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, catalog.Uint32Size)
			binary.LittleEndian.PutUint32(data, uint32(tt.value))

			v, err := tuple.NewInt32ValueFromBytes(data)
			if err != nil {
				t.Fatalf("NewInt32ValueFromBytes() error = %v, want nil", err)
			}

			if v.Value != tt.value {
				t.Errorf("Value = %d, want %d", v.Value, tt.value)
			}
		})
	}
}

func TestInt32ValueFromBytesInvalidSize(t *testing.T) {
	sizes := []int{0, 1, 2, 3, 5, 8, 16}

	for _, size := range sizes {
		t.Run("size_"+strconv.Itoa(size), func(t *testing.T) {
			data := make([]byte, size)
			v, err := tuple.NewInt32ValueFromBytes(data)
			if v != nil {
				t.Errorf("NewInt32ValueFromBytes() = %v, want nil", v)
			}
			if err != tuple.ErrInvalidValue {
				t.Errorf("error = %v, want %v", err, tuple.ErrInvalidValue)
			}
		})
	}
}

func TestInt32ValueRoundTrip(t *testing.T) {
	original := tuple.Int32Value{Value: -12345}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v, want nil", err)
	}

	restored, err := tuple.NewInt32ValueFromBytes(data)
	if err != nil {
		t.Fatalf("NewInt32ValueFromBytes() error = %v, want nil", err)
	}

	if restored.Value != original.Value {
		t.Errorf("round trip Value = %d, want %d", restored.Value, original.Value)
	}
}

func TestInt32ValueString(t *testing.T) {
	tests := []struct {
		value int32
		want  string
	}{
		{0, "0"},
		{42, "42"},
		{-1, "-1"},
		{math.MaxInt32, "2147483647"},
		{math.MinInt32, "-2147483648"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			v := tuple.Int32Value{Value: tt.value}
			got := v.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInt32ValueImplementsValue(t *testing.T) {
	var v tuple.Value = tuple.Int32Value{Value: 1}

	if v.Type() != catalog.TypeInt32Type {
		t.Errorf("Type() = %v, want %v", v.Type(), catalog.TypeInt32Type)
	}

	data, err := v.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	if len(data) != catalog.Uint32Size {
		t.Errorf("Serialize() len = %d, want %d", len(data), catalog.Uint32Size)
	}

	if v.Size() != catalog.Uint32Size {
		t.Errorf("Size() = %d, want %d", v.Size(), catalog.Uint32Size)
	}
}

func TestInt32ValueZeroValue(t *testing.T) {
	v := tuple.Int32Value{}

	if v.Value != 0 {
		t.Errorf("default Value = %d, want 0", v.Value)
	}

	if v.Type() != catalog.TypeInt32Type {
		t.Errorf("Type() = %v, want %v", v.Type(), catalog.TypeInt32Type)
	}

	data, err := v.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	got := int32(binary.LittleEndian.Uint32(data))
	if got != 0 {
		t.Errorf("serialized zero value = %d, want 0", got)
	}
}

func TestInt32ValueFromBytesNil(t *testing.T) {
	v, err := tuple.NewInt32ValueFromBytes(nil)
	if v != nil {
		t.Errorf("NewInt32ValueFromBytes(nil) = %v, want nil", v)
	}
	if err != tuple.ErrInvalidValue {
		t.Errorf("error = %v, want %v", err, tuple.ErrInvalidValue)
	}
}
