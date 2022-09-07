// Code generated by https://github.com/meerkat-io/disorder; DO NOT EDIT.
package test

import (
	"fmt"
	"github.com/meerkat-io/disorder/internal/test_data/test/sub"
	"time"
)

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

func (enum *Color) Decode(value string) error {
	if color, ok := colorEnumMap[value]; ok {
		*enum = color
		return nil
	}
	return fmt.Errorf("invalid enum value")
}

func (enum *Color) Encode() ([]byte, error) {
	name := string(*enum)
	if len(name) == 0 {
		return nil, fmt.Errorf("empty enum value")
	}
	if len(name) > 255 {
		return nil, fmt.Errorf("string length overflow. should less than 255")
	}
	data := []byte{byte(len(name))}
	if _, ok := colorEnumMap[name]; ok {
		data = append(data, []byte(name)...)
		return data, nil
	}
	return nil, fmt.Errorf("invalid enum value")
}

type Object struct {
	EnumField   *Color                    `disorder:"enum_field" json:"enum_field,omitempty"`
	ObjMap      map[string]*sub.SubObject `disorder:"obj_map" json:"obj_map,omitempty"`
	EmptyString string                    `disorder:"empty_string" json:"empty_string"`
	EmptyEnum   *Color                    `disorder:"empty_enum" json:"empty_enum,omitempty"`
	EmptyObj    *sub.SubObject            `disorder:"empty_obj" json:"empty_obj,omitempty"`
	ObjField    *sub.SubObject            `disorder:"obj_field" json:"obj_field,omitempty"`
	IntArray    []int32                   `disorder:"int_array" json:"int_array,omitempty"`
	IntMap      map[string]int32          `disorder:"int_map" json:"int_map,omitempty"`
	EmptyTime   *time.Time                `disorder:"empty_time" json:"empty_time,omitempty"`
	IntField    int32                     `disorder:"int_field" json:"int_field"`
	StringField string                    `disorder:"string_field" json:"string_field"`
	TimeField   *time.Time                `disorder:"time_field" json:"time_field,omitempty"`
	ObjArray    []*sub.SubObject          `disorder:"obj_array" json:"obj_array,omitempty"`
	EmptyArray  []int32                   `disorder:"empty_array" json:"empty_array,omitempty"`
	EmptyMap    map[string]int32          `disorder:"empty_map" json:"empty_map,omitempty"`
}

type Zero struct {
	ZeroMap   map[string]int32 `disorder:"zero_map" json:"zero_map,omitempty"`
	ZeroArray []int32          `disorder:"zero_array" json:"zero_array,omitempty"`
}
