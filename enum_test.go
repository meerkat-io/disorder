package disorder_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/meerkat-io/disorder"
	"github.com/stretchr/testify/assert"
)

func TestTypedEnum(t *testing.T) {
	c := Color{value: "color"}

	var ptr interface{} = &c
	assert.NotNil(t, ptr.(disorder.Enum))

	_, err := c.GetValue()
	assert.NotNil(t, err)

	err = c.SetValue("blue")
	assert.Nil(t, err)

	s, err := c.GetValue()
	assert.Nil(t, err)
	assert.Equal(t, s, "blue")

	v := reflect.ValueOf(ptr)
	enum := v.Interface().(disorder.Enum)
	assert.NotNil(t, enum)
}

type Color struct {
	value string
}

var (
	ColorRed   = Color{value: "red"}
	ColorGreen = Color{value: "green"}
	ColorBlue  = Color{value: "blue"}
)

var colorEnumMap = map[string]Color{
	"red":   ColorRed,
	"green": ColorGreen,
	"blue":  ColorBlue,
}

func (*Color) Enum() {}

func (enum *Color) SetValue(value string) error {
	if value == "" {
		return fmt.Errorf("empty enum value")
	}
	if len(value) > 255 {
		return fmt.Errorf("enum length overflow. should less than 255")
	}
	if _, ok := colorEnumMap[value]; ok {
		enum.value = value
		return nil
	}
	return fmt.Errorf("invalid enum value: %s", value)
}

func (enum *Color) GetValue() (string, error) {
	if len(enum.value) == 0 {
		return "", fmt.Errorf("empty enum value")
	}
	if len(enum.value) > 255 {
		return "", fmt.Errorf("enum length overflow. should less than 255")
	}
	if _, ok := colorEnumMap[enum.value]; ok {
		return enum.value, nil
	}
	return "", fmt.Errorf("invalid enum value: %s", enum.value)
}
