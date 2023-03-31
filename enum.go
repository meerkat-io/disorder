package disorder

import "fmt"

type Enum interface {
	Enum()
	FromString(enum string) error
	ToString() (string, error)
}

type EnumBase string

func (*EnumBase) Enum() {}

func (enum *EnumBase) FromString(value string) error {
	if value == "" {
		return fmt.Errorf("empty enum value")
	}
	if len(value) > 255 {
		return fmt.Errorf("enum length overflow. should less than 255")
	}
	*enum = EnumBase(value)
	return nil
}

func (enum *EnumBase) ToString() (string, error) {
	name := string(*enum)
	if len(name) == 0 {
		return "", fmt.Errorf("empty enum value")
	}
	if len(name) > 255 {
		return "", fmt.Errorf("enum length overflow. should less than 255")
	}
	return name, nil
}
