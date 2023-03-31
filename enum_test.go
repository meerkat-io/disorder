package disorder_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/meerkat-io/disorder"
	"github.com/stretchr/testify/assert"
)

func TestTypedEnum(t *testing.T) {
	c := Color("color")
	var ptr interface{} = &c
	assert.NotNil(t, ptr.(disorder.Enum))
	_, err := c.ToString()
	assert.NotNil(t, err)
	err = c.FromString("blue")
	assert.Nil(t, err)
	s, err := c.ToString()
	assert.Nil(t, err)
	assert.Equal(t, s, "blue")

	v := reflect.ValueOf(ptr)
	enum := v.Interface().(disorder.Enum)
	assert.NotNil(t, enum)
}

func TestDynamicEnum(t *testing.T) {
	c := disorder.EnumBase("color")
	var ptr interface{} = &c
	assert.NotNil(t, ptr.(disorder.Enum))
	s, err := c.ToString()
	assert.Nil(t, err)
	assert.Equal(t, s, "color")

	v := reflect.ValueOf(ptr)
	enum := v.Interface().(disorder.Enum)
	assert.NotNil(t, enum)
}

type Color string

const (
	ColorRed   = Color("red")
	ColorGreen = Color("green")
	ColorBlue  = Color("blue")
)

var colorEnumMap = map[string]Color{
	"red":   ColorRed,
	"green": ColorGreen,
	"blue":  ColorBlue,
}

func (*Color) Enum() {}

func (enum *Color) FromString(value string) error {
	if value == "" {
		return fmt.Errorf("empty enum value")
	}
	if len(value) > 255 {
		return fmt.Errorf("enum length overflow. should less than 255")
	}
	if color, ok := colorEnumMap[value]; ok {
		*enum = color
		return nil
	}
	return fmt.Errorf("invalid enum value: %s", value)
}

func (enum *Color) ToString() (string, error) {
	name := string(*enum)
	if len(name) == 0 {
		return "", fmt.Errorf("empty enum value")
	}
	if len(name) > 255 {
		return "", fmt.Errorf("enum length overflow. should less than 255")
	}
	if _, ok := colorEnumMap[name]; ok {
		return name, nil
	}
	return "", fmt.Errorf("invalid enum value: %s", name)
}
