package catalog

type DataType uint8

const (
	TypeInt32Type DataType = iota
	TypeStringType
	TypeBoolType
)

func (d DataType) IsValid() bool {
	return d >= TypeInt32Type && d <= TypeBoolType
}
