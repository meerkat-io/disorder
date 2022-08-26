// Code generated by https://github.com/meerkat-io/disorder; DO NOT EDIT.
package test

import (
	"fmt"
	"github.com/meerkat-io/disorder/internal/test_data/test/sub"
	"time"
)

type Color int

const (
	ColorRed Color = iota
	ColorGreen
	ColorBlue
)

func (*Color) Enum() {}

func (enum *Color) FromString(value string) error {
	switch {
	case value == "red":
		*enum = ColorRed
		return nil
	case value == "green":
		*enum = ColorGreen
		return nil
	case value == "blue":
		*enum = ColorBlue
		return nil
	}
	return fmt.Errorf("invalid enum value")
}

func (enum *Color) ToString() (string, error) {
	switch *enum {
	case ColorRed:
		return "red", nil
	case ColorGreen:
		return "green", nil
	case ColorBlue:
		return "blue", nil
	default:
		return "", fmt.Errorf("invalid enum value")
	}
}

type Object struct {
	IntField    int32                     `disorder:"int_field"`
	StringField string                    `disorder:"string_field"`
	IntMap      map[string]int32          `disorder:"int_map"`
	NullObj     *sub.SubObject            `disorder:"null_obj"`
	EnumField   *Color                    `disorder:"enum_field"`
	IntArray    []int32                   `disorder:"int_array"`
	ObjArray    []*sub.SubObject          `disorder:"obj_array"`
	ObjMap      map[string]*sub.SubObject `disorder:"obj_map"`
	Time        *time.Time                `disorder:"time"`
}
