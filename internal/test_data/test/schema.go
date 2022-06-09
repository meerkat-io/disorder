// Code generated by https://github.com/meerkat-lib/disorder; DO NOT EDIT.
package test

import (
	"github.com/meerkat-lib/disorder/internal/test_data/test/sub"
)

type Color string

func (Color) Enum() {}

const (
	ColorRed   Color = "red"
	ColorGreen Color = "green"
	ColorBlue  Color = "blue"
)

type Animal string

func (Animal) Enum() {}

const (
	AnimalDog Animal = "dog"
	AnimalCat Animal = "cat"
)

type Object struct {
	EnumField   Color                     `disorder:"enum_field"`
	IntArray    []int32                   `disorder:"int_array"`
	IntMap      map[string]int32          `disorder:"int_map"`
	ObjArray    []*sub.SubObject          `disorder:"obj_array"`
	ObjMap      map[string]*sub.SubObject `disorder:"obj_map"`
	IntField    int32                     `disorder:"int_field"`
	StringField string                    `disorder:"string_field"`
}

type AnotherObject struct {
	Value int32 `disorder:"value"`
}
