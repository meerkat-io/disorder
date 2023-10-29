package disorder_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/meerkat-io/disorder"
	"github.com/stretchr/testify/assert"
)

func TestEnum(t *testing.T) {
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

func TestPrimaryTypes(t *testing.T) {
	var b bool
	data, err := disorder.Marshal(true)
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &b)
	assert.Nil(t, err)
	assert.Equal(t, true, b)

	var i int32
	data, err = disorder.Marshal(int32(-123))
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &i)
	assert.Nil(t, err)
	assert.Equal(t, int32(-123), i)

	var l int64
	data, err = disorder.Marshal(int64(-456))
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &l)
	assert.Nil(t, err)
	assert.Equal(t, int64(-456), l)

	var f float32
	data, err = disorder.Marshal(float32(-3.14))
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &f)
	assert.Nil(t, err)
	assert.Equal(t, float32(-3.14), f)

	var d float64
	data, err = disorder.Marshal(float64(-6.28))
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &d)
	assert.Nil(t, err)
	assert.Equal(t, float64(-6.28), d)

	var bytes []byte
	data, err = disorder.Marshal([]byte{1, 2, 3})
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &bytes)
	assert.Nil(t, err)
	assert.Equal(t, []byte{1, 2, 3}, bytes)
}

func TestUnsupportedEncodeTypes(t *testing.T) {
	_, err := disorder.Marshal(int(123))
	assert.NotNil(t, err)

	_, err = disorder.Marshal(nil)
	assert.NotNil(t, err)
}

func TestUtilTypes(t *testing.T) {
	var s string
	data, err := disorder.Marshal("hello world")
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &s)
	assert.Nil(t, err)
	assert.Equal(t, "hello world", s)

	var now time.Time = time.UnixMilli(time.Now().UnixMilli())
	var result time.Time
	data, err = disorder.Marshal(&now)
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &result)
	assert.Nil(t, err)
	assert.Equal(t, result, now)

	var c Color
	data, err = disorder.Marshal(&ColorBlue)
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &c)
	assert.Nil(t, err)
	assert.Equal(t, ColorBlue, c)
}

func TestContainerTypes(t *testing.T) {
	var list []string
	data, err := disorder.Marshal([]string{"hello", "world"})
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &list)
	assert.Nil(t, err)
	assert.Equal(t, []string{"hello", "world"}, list)

	var set map[string]string
	data, err = disorder.Marshal(map[string]string{"hello": "world"})
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &set)
	assert.Nil(t, err)
	assert.Equal(t, map[string]string{"hello": "world"}, set)

	var complex map[string][]*Color
	data, err = disorder.Marshal(map[string][]*Color{"hello": {&ColorBlue}})
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &complex)
	assert.Nil(t, err)
	assert.Equal(t, map[string][]*Color{"hello": {&ColorBlue}}, complex)
}

func TestStructTypes(t *testing.T) {
	number0 := Number{
		Value: 123,
	}
	json0, err := json.Marshal(&number0)
	assert.Nil(t, err)
	data0, err := disorder.Marshal(number0)
	assert.Nil(t, err)

	var number1 Number
	err = disorder.Unmarshal(data0, &number1)
	assert.Nil(t, err)
	json1, err := json.Marshal(number1)
	assert.Nil(t, err)

	assert.Equal(t, number0, number1)
	assert.JSONEq(t, string(json0), string(json1))

	timestamp := time.UnixMilli(time.Now().UnixMilli())
	object0 := Object{
		BooleanField: true,
		IntField:     123,
		StringField:  "foo",
		BytesFields:  []byte{7, 8, 9},
		EnumField:    &ColorBlue,
		TimeField:    &timestamp,
		ObjField: &NumberWrapper{
			Value: &Number{
				Value: 789,
			},
		},
		IntArray: []int32{1, 2, 3},
		IntMap: map[string]int32{
			"4": 4,
			"5": 5,
			"6": 6,
		},
		ObjArray: []*NumberWrapper{{Value: &Number{
			Value: 789,
		}}},
		ObjMap: map[string]*NumberWrapper{
			"789": {Value: &Number{
				Value: 789,
			}},
		},
		Nested: map[string]map[string][][]map[string]*Color{
			"key0": {
				"key1": {
					{
						{
							"key2": &ColorBlue,
						},
					},
				},
			},
		},
	}
	json0, err = json.Marshal(object0)
	assert.Nil(t, err)
	data0, err = disorder.Marshal(&object0)
	assert.Nil(t, err)

	var object1 Object
	err = disorder.Unmarshal(data0, &object1)
	assert.Nil(t, err)
	json1, err = json.Marshal(object1)
	assert.Nil(t, err)

	assert.Equal(t, object0, object1)
	assert.JSONEq(t, string(json0), string(json1))
}

func TestSkipStructFields(t *testing.T) {
	timestamp := time.UnixMilli(time.Now().UnixMilli())
	object0 := Object{
		BooleanField: true,
		IntField:     123,
		StringField:  "foo",
		BytesFields:  []byte{7, 8, 9},
		EnumField:    &ColorBlue,
		TimeField:    &timestamp,
		ObjField: &NumberWrapper{
			Value: &Number{
				Value: 789,
			},
		},
		IntArray: []int32{1, 2, 3},
		IntMap: map[string]int32{
			"4": 4,
			"5": 5,
			"6": 6,
		},
		ObjArray: []*NumberWrapper{{Value: &Number{
			Value: 789,
		}}},
		ObjMap: map[string]*NumberWrapper{
			"789": {Value: &Number{
				Value: 789,
			}},
		},
		Nested: map[string]map[string][][]map[string]*Color{
			"key0": {
				"key1": {
					{
						{
							"key2": &ColorBlue,
						},
					},
				},
			},
		},
		EmptyString: "not empty",
	}
	data0, err := disorder.Marshal(&object0)
	assert.Nil(t, err)

	var object1 SkipObject
	err = disorder.Unmarshal(data0, &object1)
	assert.Nil(t, err)

	assert.Equal(t, object0.EmptyString, object1.EmptyString)
}

func TestZero(t *testing.T) {
	object0 := Zero{
		ZeroArray: []int32{},
		ZeroMap:   map[string]int32{},
	}
	assert.NotNil(t, object0.ZeroArray)
	assert.NotNil(t, object0.ZeroMap)
	assert.Equal(t, 0, len(object0.ZeroArray))
	assert.Equal(t, 0, len(object0.ZeroMap))
	data0, err := disorder.Marshal(&object0)
	assert.Nil(t, err)

	object1 := Zero{}
	err = disorder.Unmarshal(data0, &object1)
	assert.Nil(t, err)
	assert.NotNil(t, object1.ZeroArray)
	assert.NotNil(t, object1.ZeroMap)
	assert.Equal(t, 0, len(object1.ZeroArray))
	assert.Equal(t, 0, len(object1.ZeroMap))
}
