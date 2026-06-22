package tuple_test

import (
	"math"
	"testing"

	"db-engine-v1/internal/catalog"
	"db-engine-v1/internal/storage/tuple"
)

func TestNewTuple(t *testing.T) {
	tup := tuple.Tuple{
		Values: []tuple.Value{
			tuple.Int32Value{Value: 42},
			tuple.StringValue{Value: "hello"},
			tuple.BoolValue{Value: true},
		},
	}

	if len(tup.Values) != 3 {
		t.Fatalf("len(Values) = %d, want 3", len(tup.Values))
	}
}

func TestTupleSize(t *testing.T) {
	tup := tuple.Tuple{
		Values: []tuple.Value{
			tuple.Int32Value{Value: 1},
			tuple.StringValue{Value: "abc"},
			tuple.BoolValue{Value: false},
		},
	}

	expected := uint16(catalog.Uint32Size + catalog.StringLengthSize + 3 + catalog.BoolSize)
	if tup.Size() != expected {
		t.Errorf("Size() = %d, want %d", tup.Size(), expected)
	}
}

func TestTupleSizeEmpty(t *testing.T) {
	tup := tuple.Tuple{Values: []tuple.Value{}}
	if tup.Size() != 0 {
		t.Errorf("Size() = %d, want 0", tup.Size())
	}
}

func TestTupleSerialize(t *testing.T) {
	tup := tuple.Tuple{
		Values: []tuple.Value{
			tuple.Int32Value{Value: 42},
			tuple.StringValue{Value: "hi"},
		},
	}

	data, err := tup.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	expectedSize := catalog.Uint32Size + catalog.StringLengthSize + 2
	if len(data) != expectedSize {
		t.Fatalf("len(data) = %d, want %d", len(data), expectedSize)
	}
}

func TestNewTupleFromBytesInt32(t *testing.T) {
	tup := tuple.Tuple{Values: []tuple.Value{tuple.Int32Value{Value: -100}}}
	data, err := tup.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	columns := []catalog.Column{{Name: "id", Type: catalog.TypeInt32Type}}
	restored, err := tuple.NewTupleFromBytes(data, columns)
	if err != nil {
		t.Fatalf("NewTupleFromBytes() error = %v", err)
	}

	if len(restored.Values) != 1 {
		t.Fatalf("len(Values) = %d, want 1", len(restored.Values))
	}

	v, ok := restored.Values[0].(*tuple.Int32Value)
	if !ok {
		t.Fatalf("Values[0] type = %T, want *Int32Value", restored.Values[0])
	}
	if v.Value != -100 {
		t.Errorf("Values[0].Value = %d, want -100", v.Value)
	}
}

func TestNewTupleFromBytesString(t *testing.T) {
	tup := tuple.Tuple{Values: []tuple.Value{tuple.StringValue{Value: "café ñoño"}}}
	data, err := tup.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	columns := []catalog.Column{{Name: "name", Type: catalog.TypeStringType}}
	restored, err := tuple.NewTupleFromBytes(data, columns)
	if err != nil {
		t.Fatalf("NewTupleFromBytes() error = %v", err)
	}

	v, ok := restored.Values[0].(*tuple.StringValue)
	if !ok {
		t.Fatalf("Values[0] type = %T, want *StringValue", restored.Values[0])
	}
	if v.Value != "café ñoño" {
		t.Errorf("Values[0].Value = %q, want %q", v.Value, "café ñoño")
	}
}

func TestNewTupleFromBytesBool(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		tup := tuple.Tuple{Values: []tuple.Value{tuple.BoolValue{Value: true}}}
		data, _ := tup.Serialize()
		columns := []catalog.Column{{Name: "active", Type: catalog.TypeBoolType}}
		restored, err := tuple.NewTupleFromBytes(data, columns)
		if err != nil {
			t.Fatalf("NewTupleFromBytes() error = %v", err)
		}
		v, ok := restored.Values[0].(*tuple.BoolValue)
		if !ok {
			t.Fatalf("Values[0] type = %T, want *BoolValue", restored.Values[0])
		}
		if v.Value != true {
			t.Errorf("Value = %v, want true", v.Value)
		}
	})

	t.Run("false", func(t *testing.T) {
		tup := tuple.Tuple{Values: []tuple.Value{tuple.BoolValue{Value: false}}}
		data, _ := tup.Serialize()
		columns := []catalog.Column{{Name: "active", Type: catalog.TypeBoolType}}
		restored, err := tuple.NewTupleFromBytes(data, columns)
		if err != nil {
			t.Fatalf("NewTupleFromBytes() error = %v", err)
		}
		v, ok := restored.Values[0].(*tuple.BoolValue)
		if !ok {
			t.Fatalf("Values[0] type = %T, want *BoolValue", restored.Values[0])
		}
		if v.Value != false {
			t.Errorf("Value = %v, want false", v.Value)
		}
	})
}

func TestNewTupleFromBytesMixedTypes(t *testing.T) {
	tup := tuple.Tuple{
		Values: []tuple.Value{
			tuple.Int32Value{Value: 1000},
			tuple.StringValue{Value: "hello"},
			tuple.BoolValue{Value: true},
			tuple.Int32Value{Value: -999},
		},
	}

	data, err := tup.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	columns := []catalog.Column{
		{Name: "id", Type: catalog.TypeInt32Type},
		{Name: "name", Type: catalog.TypeStringType},
		{Name: "active", Type: catalog.TypeBoolType},
		{Name: "count", Type: catalog.TypeInt32Type},
	}

	restored, err := tuple.NewTupleFromBytes(data, columns)
	if err != nil {
		t.Fatalf("NewTupleFromBytes() error = %v", err)
	}

	if len(restored.Values) != 4 {
		t.Fatalf("len(Values) = %d, want 4", len(restored.Values))
	}

	v0, ok0 := restored.Values[0].(*tuple.Int32Value)
	if !ok0 || v0.Value != 1000 {
		t.Errorf("Values[0] = %+v, want Int32Value(1000)", restored.Values[0])
	}

	v1, ok1 := restored.Values[1].(*tuple.StringValue)
	if !ok1 || v1.Value != "hello" {
		t.Errorf("Values[1] = %+v, want StringValue(hello)", restored.Values[1])
	}

	v2, ok2 := restored.Values[2].(*tuple.BoolValue)
	if !ok2 || v2.Value != true {
		t.Errorf("Values[2] = %+v, want BoolValue(true)", restored.Values[2])
	}

	v3, ok3 := restored.Values[3].(*tuple.Int32Value)
	if !ok3 || v3.Value != -999 {
		t.Errorf("Values[3] = %+v, want Int32Value(-999)", restored.Values[3])
	}
}

func TestNewTupleFromBytesEmptyColumns(t *testing.T) {
	_, err := tuple.NewTupleFromBytes([]byte{1, 2, 3}, []catalog.Column{})
	if err != tuple.ErrInvalidTuple {
		t.Errorf("error = %v, want %v", err, tuple.ErrInvalidTuple)
	}
}

func TestNewTupleFromBytesExtraData(t *testing.T) {
	tup := tuple.Tuple{Values: []tuple.Value{tuple.Int32Value{Value: 42}}}
	data, _ := tup.Serialize()

	data = append(data, 0xFF)

	columns := []catalog.Column{{Name: "id", Type: catalog.TypeInt32Type}}
	_, err := tuple.NewTupleFromBytes(data, columns)
	if err != tuple.ErrInvalidTuple {
		t.Errorf("error = %v, want %v", err, tuple.ErrInvalidTuple)
	}
}

func TestNewTupleFromBytesInvalidDataType(t *testing.T) {
	columns := []catalog.Column{{Name: "bad", Type: catalog.DataType(99)}}
	_, err := tuple.NewTupleFromBytes([]byte{0, 0, 0, 0}, columns)
	if err != tuple.ErrInvalidDataType {
		t.Errorf("error = %v, want %v", err, tuple.ErrInvalidDataType)
	}
}

func TestNewTupleFromBytesBadStringData(t *testing.T) {
	data := []byte{100, 0, 'a', 'b'}
	columns := []catalog.Column{{Name: "name", Type: catalog.TypeStringType}}
	_, err := tuple.NewTupleFromBytes(data, columns)
	if err != tuple.ErrInvalidValue {
		t.Errorf("error = %v, want %v", err, tuple.ErrInvalidValue)
	}
}

func TestNewTupleFromBytesEmptyData(t *testing.T) {
	columns := []catalog.Column{{Name: "id", Type: catalog.TypeInt32Type}}
	_, err := tuple.NewTupleFromBytes([]byte{}, columns)
	if err != tuple.ErrInvalidValue {
		t.Errorf("error = %v, want %v", err, tuple.ErrInvalidValue)
	}
}

func TestTupleRoundTrip(t *testing.T) {
	t.Run("all_types", func(t *testing.T) {
		original := tuple.Tuple{
			Values: []tuple.Value{
				tuple.Int32Value{Value: math.MaxInt32},
				tuple.StringValue{Value: "áéíóú ñ 😊"},
				tuple.BoolValue{Value: true},
				tuple.Int32Value{Value: math.MinInt32},
				tuple.StringValue{Value: ""},
				tuple.BoolValue{Value: false},
			},
		}

		data, err := original.Serialize()
		if err != nil {
			t.Fatalf("Serialize() error = %v", err)
		}

		columns := []catalog.Column{
			{Name: "a", Type: catalog.TypeInt32Type},
			{Name: "b", Type: catalog.TypeStringType},
			{Name: "c", Type: catalog.TypeBoolType},
			{Name: "d", Type: catalog.TypeInt32Type},
			{Name: "e", Type: catalog.TypeStringType},
			{Name: "f", Type: catalog.TypeBoolType},
		}

		restored, err := tuple.NewTupleFromBytes(data, columns)
		if err != nil {
			t.Fatalf("NewTupleFromBytes() error = %v", err)
		}

		if len(restored.Values) != 6 {
			t.Fatalf("len(Values) = %d, want 6", len(restored.Values))
		}

		if v := restored.Values[0].(*tuple.Int32Value).Value; v != math.MaxInt32 {
			t.Errorf("Values[0] = %d, want %d", v, math.MaxInt32)
		}
		if v := restored.Values[1].(*tuple.StringValue).Value; v != "áéíóú ñ 😊" {
			t.Errorf("Values[1] = %q, want %q", v, "áéíóú ñ 😊")
		}
		if v := restored.Values[2].(*tuple.BoolValue).Value; v != true {
			t.Errorf("Values[2] = %v, want true", v)
		}
		if v := restored.Values[3].(*tuple.Int32Value).Value; v != math.MinInt32 {
			t.Errorf("Values[3] = %d, want %d", v, math.MinInt32)
		}
		if v := restored.Values[4].(*tuple.StringValue).Value; v != "" {
			t.Errorf("Values[4] = %q, want empty", v)
		}
		if v := restored.Values[5].(*tuple.BoolValue).Value; v != false {
			t.Errorf("Values[5] = %v, want false", v)
		}
	})

	t.Run("single_int32", func(t *testing.T) {
		original := tuple.Tuple{Values: []tuple.Value{tuple.Int32Value{Value: 12345}}}
		data, _ := original.Serialize()
		columns := []catalog.Column{{Name: "id", Type: catalog.TypeInt32Type}}
		restored, err := tuple.NewTupleFromBytes(data, columns)
		if err != nil {
			t.Fatalf("NewTupleFromBytes() error = %v", err)
		}
		if restored.Values[0].(*tuple.Int32Value).Value != 12345 {
			t.Errorf("Values[0] = %d, want 12345", restored.Values[0].(*tuple.Int32Value).Value)
		}
	})

	t.Run("single_string", func(t *testing.T) {
		original := tuple.Tuple{Values: []tuple.Value{tuple.StringValue{Value: "über cool"}}}
		data, _ := original.Serialize()
		columns := []catalog.Column{{Name: "name", Type: catalog.TypeStringType}}
		restored, err := tuple.NewTupleFromBytes(data, columns)
		if err != nil {
			t.Fatalf("NewTupleFromBytes() error = %v", err)
		}
		if restored.Values[0].(*tuple.StringValue).Value != "über cool" {
			t.Errorf("Values[0] = %q, want %q", restored.Values[0].(*tuple.StringValue).Value, "über cool")
		}
	})

	t.Run("single_bool", func(t *testing.T) {
		original := tuple.Tuple{Values: []tuple.Value{tuple.BoolValue{Value: true}}}
		data, _ := original.Serialize()
		columns := []catalog.Column{{Name: "active", Type: catalog.TypeBoolType}}
		restored, err := tuple.NewTupleFromBytes(data, columns)
		if err != nil {
			t.Fatalf("NewTupleFromBytes() error = %v", err)
		}
		if restored.Values[0].(*tuple.BoolValue).Value != true {
			t.Errorf("Values[0] = %v, want true", restored.Values[0].(*tuple.BoolValue).Value)
		}
	})
}

func TestNewTupleFromBytesPartialData(t *testing.T) {
	data := []byte{0x01, 0x00}
	columns := []catalog.Column{{Name: "id", Type: catalog.TypeInt32Type}}
	_, err := tuple.NewTupleFromBytes(data, columns)
	if err != tuple.ErrInvalidValue {
		t.Errorf("error = %v, want %v", err, tuple.ErrInvalidValue)
	}
}

func TestNewTupleFromBytesInvalidBool(t *testing.T) {
	data := []byte{0xFF}
	columns := []catalog.Column{{Name: "active", Type: catalog.TypeBoolType}}
	_, err := tuple.NewTupleFromBytes(data, columns)
	if err != tuple.ErrInvalidValue {
		t.Errorf("error = %v, want %v", err, tuple.ErrInvalidValue)
	}
}

func TestTupleSizeWithAllTypes(t *testing.T) {
	tup := tuple.Tuple{
		Values: []tuple.Value{
			tuple.Int32Value{Value: 0},
			tuple.StringValue{Value: "hello"},
			tuple.BoolValue{Value: true},
		},
	}

	expected := catalog.Uint32Size + catalog.StringLengthSize + 5 + catalog.BoolSize
	if tup.Size() != uint16(expected) {
		t.Errorf("Size() = %d, want %d", tup.Size(), expected)
	}

	data, _ := tup.Serialize()
	if len(data) != expected {
		t.Errorf("Serialize() len = %d, want %d", len(data), expected)
	}
}

func TestNewTupleFromBytesEmptyStringColumn(t *testing.T) {
	tup := tuple.Tuple{
		Values: []tuple.Value{
			tuple.StringValue{Value: ""},
		},
	}

	data, err := tup.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	columns := []catalog.Column{{Name: "note", Type: catalog.TypeStringType}}
	restored, err := tuple.NewTupleFromBytes(data, columns)
	if err != nil {
		t.Fatalf("NewTupleFromBytes() error = %v", err)
	}

	v := restored.Values[0].(*tuple.StringValue)
	if v.Value != "" {
		t.Errorf("Value = %q, want empty", v.Value)
	}
}

func TestNewTupleFromBytesMultipleStringsUTF8(t *testing.T) {
	tup := tuple.Tuple{
		Values: []tuple.Value{
			tuple.StringValue{Value: "áéíóú"},
			tuple.StringValue{Value: "ñññ"},
			tuple.StringValue{Value: "😊😊"},
		},
	}

	data, err := tup.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	columns := []catalog.Column{
		{Name: "a", Type: catalog.TypeStringType},
		{Name: "b", Type: catalog.TypeStringType},
		{Name: "c", Type: catalog.TypeStringType},
	}

	restored, err := tuple.NewTupleFromBytes(data, columns)
	if err != nil {
		t.Fatalf("NewTupleFromBytes() error = %v", err)
	}

	if v := restored.Values[0].(*tuple.StringValue).Value; v != "áéíóú" {
		t.Errorf("Values[0] = %q, want %q", v, "áéíóú")
	}
	if v := restored.Values[1].(*tuple.StringValue).Value; v != "ñññ" {
		t.Errorf("Values[1] = %q, want %q", v, "ñññ")
	}
	if v := restored.Values[2].(*tuple.StringValue).Value; v != "😊😊" {
		t.Errorf("Values[2] = %q, want %q", v, "😊😊")
	}
}
