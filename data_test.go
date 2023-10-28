package disorder_test

import (
	"fmt"
	"time"
)

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

type Object struct {
	BooleanField bool                                        `disorder:"boolean_field" json:"boolean_field"`
	IntField     int32                                       `disorder:"int_field" json:"int_field"`
	StringField  string                                      `disorder:"string_field" json:"string_field"`
	BytesFields  []byte                                      `disorder:"bytes_fields" json:"bytes_fields"`
	EnumField    *Color                                      `disorder:"enum_field" json:"enum_field,omitempty"`
	TimeField    *time.Time                                  `disorder:"time_field" json:"time_field,omitempty"`
	ObjField     *NumberWrapper                              `disorder:"obj_field" json:"obj_field,omitempty"`
	IntArray     []int32                                     `disorder:"int_array" json:"int_array,omitempty"`
	IntMap       map[string]int32                            `disorder:"int_map" json:"int_map,omitempty"`
	ObjArray     []*NumberWrapper                            `disorder:"obj_array" json:"obj_array,omitempty"`
	ObjMap       map[string]*NumberWrapper                   `disorder:"obj_map" json:"obj_map,omitempty"`
	EmptyString  string                                      `disorder:"empty_string" json:"empty_string"`
	EmptyEnum    *Color                                      `disorder:"empty_enum" json:"empty_enum,omitempty"`
	EmptyTime    *time.Time                                  `disorder:"empty_time" json:"empty_time,omitempty"`
	EmptyObj     *NumberWrapper                              `disorder:"empty_obj" json:"empty_obj,omitempty"`
	EmptyArray   []int32                                     `disorder:"empty_array" json:"empty_array,omitempty"`
	EmptyMap     map[string]int32                            `disorder:"empty_map" json:"empty_map,omitempty"`
	Nested       map[string]map[string][][]map[string]*Color `disorder:"nested" json:"nested,omitempty"`
}

type Number struct {
	Value int32 `disorder:"value" json:"value"`
}

type NumberWrapper struct {
	Value *Number `disorder:"value" json:"value,omitempty"`
}

type SkipObject struct {
	EmptyString string `disorder:"empty_string" json:"empty_string"`
}
