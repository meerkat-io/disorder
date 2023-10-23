package disorder

import "fmt"

type Enum interface {
	Enum()
	GetValue() (string, error)
	SetValue(enum string) error
}

type EnumBase string

func (*EnumBase) Enum() {}

func (enum *EnumBase) GetValue() (string, error) {
	name := string(*enum)
	if len(name) == 0 {
		return "", fmt.Errorf("empty enum value")
	}
	if len(name) > 255 {
		return "", fmt.Errorf("enum length overflow. should less than 255")
	}
	return name, nil
}

func (enum *EnumBase) SetValue(value string) error {
	if value == "" {
		return fmt.Errorf("empty enum value")
	}
	if len(value) > 255 {
		return fmt.Errorf("enum length overflow. should less than 255")
	}
	*enum = EnumBase(value)
	return nil
}
