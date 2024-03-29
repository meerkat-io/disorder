package internal_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/meerkat-io/disorder"
	"github.com/meerkat-io/disorder/internal/generator/golang"
	"github.com/meerkat-io/disorder/internal/loader"
	"github.com/meerkat-io/disorder/internal/test_data/test"
	"github.com/meerkat-io/disorder/internal/test_data/test/sub"
	"github.com/stretchr/testify/assert"
)

func TestLoadSchemaFile(t *testing.T) {
	loader := loader.NewLoader()
	files, qualifiedPath, err := loader.Load("./internal/test_data/schema.yaml")
	assert.Nil(t, err)

	generator := golang.NewGoGenerator()
	err = generator.Generate("./internal", files, qualifiedPath)
	assert.Nil(t, err)

	fmt.Printf("%v\n", err)
}

func TestPrimaryType(t *testing.T) {
	var i0 int32 = 123
	var i1 int32
	data, err := disorder.Marshal(i0)
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &i1)
	assert.Nil(t, err)
	assert.Equal(t, i0, i1)
	var i2 interface{}
	err = disorder.Unmarshal(data, &i2)
	assert.Nil(t, err)
	assert.Equal(t, i0, i2)
	var i3 interface{}
	data, err = disorder.Marshal(i2)
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &i3)
	assert.Nil(t, err)
	assert.Equal(t, i0, i3)

	bytes0 := []byte{1, 2, 3, 4, 5}
	data, err = disorder.Marshal(bytes0)
	assert.Nil(t, err)
	bytes1 := []byte{}
	err = disorder.Unmarshal(data, &bytes1)
	assert.Nil(t, err)
	assert.Equal(t, bytes0, bytes1)
	var bytes2 interface{}
	err = disorder.Unmarshal(data, &bytes2)
	assert.Nil(t, err)
	assert.Equal(t, bytes0, bytes2)
	var bytes3 interface{}
	data, err = disorder.Marshal(bytes2)
	assert.Nil(t, err)
	err = disorder.Unmarshal(data, &bytes3)
	assert.Nil(t, err)
	assert.Equal(t, bytes0, bytes3)
}

func TestObjectType(t *testing.T) {
	object0 := sub.Number{
		Value: 123,
	}

	json0, err := json.Marshal(&object0)
	assert.Nil(t, err)

	data0, err := disorder.Marshal(&object0)
	assert.Nil(t, err)

	var object1 interface{}
	err = disorder.Unmarshal(data0, &object1)
	assert.Nil(t, err)
	json1, err := json.Marshal(object1)
	assert.Nil(t, err)

	data1, err := disorder.Marshal(&object1)
	assert.Nil(t, err)

	var object2 interface{}
	err = disorder.Unmarshal(data1, &object2)
	assert.Nil(t, err)
	json2, err := json.Marshal(object2)
	assert.Nil(t, err)

	data2, err := disorder.Marshal(&object2)
	assert.Nil(t, err)

	object3 := sub.Number{}
	err = disorder.Unmarshal(data2, &object3)
	assert.Nil(t, err)
	json3, err := json.Marshal(object3)
	assert.Nil(t, err)

	assert.Equal(t, object0, object3)
	assert.Equal(t, object1, object2)
	assert.JSONEq(t, string(json0), string(json1))
	assert.JSONEq(t, string(json0), string(json2))
	assert.JSONEq(t, string(json0), string(json3))
}

func TestAllTypes(t *testing.T) {
	timestamp := time.Unix(time.Now().Unix(), 0)
	color := test.ColorBlue
	object0 := test.Object{
		IntField:    123,
		StringField: "foo",
		BytesFields: []byte{7, 8, 9},
		EnumField:   &color,
		TimeField:   &timestamp,
		ObjField: &sub.NumberWrapper{
			Value: &sub.Number{
				Value: 789,
			},
		},
		IntArray: []int32{1, 2, 3},
		IntMap: map[string]int32{
			"4": 4,
			"5": 5,
			"6": 6,
		},
		ObjArray: []*sub.NumberWrapper{{Value: &sub.Number{
			Value: 789,
		}}},
		ObjMap: map[string]*sub.NumberWrapper{
			"789": {Value: &sub.Number{
				Value: 789,
			}},
		},
		Nested: map[string]map[string][][]map[string]*test.Color{
			"key0": {
				"key1": {
					{
						{
							"key2": &color,
						},
					},
				},
			},
		},
	}
	fmt.Printf("%v\n", object0)
	json0, err := json.Marshal(object0)
	assert.Nil(t, err)

	data0, err := disorder.Marshal(&object0)
	assert.Nil(t, err)

	var object1 interface{}
	err = disorder.Unmarshal(data0, &object1)
	assert.Nil(t, err)
	json1, err := json.Marshal(object1)
	assert.Nil(t, err)

	data1, err := disorder.Marshal(&object1)
	assert.Nil(t, err)

	var object2 interface{}
	err = disorder.Unmarshal(data1, &object2)
	assert.Nil(t, err)
	json2, err := json.Marshal(object2)
	assert.Nil(t, err)

	data2, err := disorder.Marshal(&object2)
	assert.Nil(t, err)

	object3 := test.Object{}
	err = disorder.Unmarshal(data2, &object3)
	assert.Nil(t, err)
	json3, err := json.Marshal(object3)
	assert.Nil(t, err)

	assert.JSONEq(t, string(json0), string(json1))
	assert.JSONEq(t, string(json0), string(json2))
	assert.JSONEq(t, string(json0), string(json3))

	miniObject := MiniObject{}
	err = disorder.Unmarshal(data0, &miniObject)
	assert.Nil(t, err)
	assert.Equal(t, object0.IntField, miniObject.IntField)
}

func TestZero(t *testing.T) {
	object0 := test.Zero{
		ZeroArray: []int32{},
		ZeroMap:   map[string]int32{},
	}
	assert.NotNil(t, object0.ZeroArray)
	assert.NotNil(t, object0.ZeroMap)
	assert.Equal(t, 0, len(object0.ZeroArray))
	assert.Equal(t, 0, len(object0.ZeroMap))

	data0, err := disorder.Marshal(&object0)
	assert.Nil(t, err)
	var object1 interface{}
	err = disorder.Unmarshal(data0, &object1)
	assert.Nil(t, err)
	assert.NotNil(t, object1.(map[string]interface{})["zero_array"])
	assert.NotNil(t, object1.(map[string]interface{})["zero_map"])

	data1, err := disorder.Marshal(&object1)
	assert.Nil(t, err)
	var object2 interface{}
	err = disorder.Unmarshal(data1, &object2)
	assert.Nil(t, err)
	assert.NotNil(t, object2.(map[string]interface{})["zero_array"])
	assert.NotNil(t, object2.(map[string]interface{})["zero_map"])

	data2, err := disorder.Marshal(&object2)
	assert.Nil(t, err)
	object3 := test.Zero{}
	err = disorder.Unmarshal(data2, &object3)
	assert.Nil(t, err)
	assert.NotNil(t, object3.ZeroArray)
	assert.NotNil(t, object3.ZeroMap)
	assert.Equal(t, 0, len(object3.ZeroArray))
	assert.Equal(t, 0, len(object3.ZeroMap))
}

/*
func TestRpcMath(t *testing.T) {
	s := rpc.NewServer()
	sub.RegisterMathServiceServer(s, &testService{})
	err := s.Listen(":9999")
	assert.Nil(t, err)

	c := sub.NewMathServiceClient(rpc.NewClient("localhost:9999"))
	result, rpcErr := c.Increase(rpc.NewContext(), int32(17))
	assert.Nil(t, rpcErr)
	assert.Equal(t, int32(18), result)

	_, rpcErr = c.Increase(rpc.NewContext(), int32(9))
	assert.NotNil(t, rpcErr)
	fmt.Println(rpcErr.Error)
}*/

/*
func TestRpcPrimary(t *testing.T) {
	s := rpc.NewServer()
	test.RegisterPrimaryServiceServer(s, &testService{})
	_ = s.Listen(":8888")

	c := test.NewPrimaryServiceClient(rpc.NewClient("localhost:8888"))

	timestamp := time.Unix(time.Now().Unix(), 0)
	color := test.ColorRed
	result1, rpcErr := c.PrintObject(rpc.NewContext(), &test.Object{
		IntField:    123,
		TimeField:   &timestamp,
		StringField: "foo",
		EnumField:   &color,
		IntArray:    []int32{1, 2, 3},
		IntMap: map[string]int32{
			"1": 1,
			"2": 2,
			"3": 3,
		},
		ObjArray: []*sub.SubObject{{Value: 123}},
		ObjMap: map[string]*sub.SubObject{
			"foo": {Value: 123},
		},
	})
	assert.Nil(t, rpcErr)
	assert.Equal(t, int32(456), result1.IntField)
	assert.Equal(t, "bar", result1.StringField)
	assert.Equal(t, test.ColorGreen, *result1.EnumField)
	assert.Equal(t, int32(4), result1.IntArray[0])
	assert.Equal(t, int32(4), result1.IntMap["4"])
	assert.Equal(t, int32(456), result1.ObjArray[0].Value)
	assert.Equal(t, int32(456), result1.ObjMap["bar"].Value)

	result2, rpcErr := c.PrintSubObject(rpc.NewContext(), &sub.SubObject{
		Value: 123,
	})
	assert.Nil(t, rpcErr)
	assert.Equal(t, int32(456), result2.Value)

	result3, rpcErr := c.PrintTime(rpc.NewContext(), &timestamp)
	assert.Nil(t, rpcErr)
	fmt.Printf("output time: %v", *result3)

	result4, rpcErr := c.PrintArray(rpc.NewContext(), []int32{1, 2, 3})
	assert.Nil(t, rpcErr)
	assert.Equal(t, 3, len(result4))
	assert.Equal(t, int32(4), result4[0])

	result5, rpcErr := c.PrintEnum(rpc.NewContext(), &color)
	assert.Nil(t, rpcErr)
	assert.Equal(t, test.ColorGreen, *result5)

	result6, rpcErr := c.PrintMap(rpc.NewContext(), map[string]string{
		"foo": "bar",
	})
	assert.Nil(t, rpcErr)
	assert.Equal(t, "foo", result6["bar"])

	result7, rpcErr := c.PrintNested(rpc.NewContext(), map[string]map[string][][]map[string]*test.Color{
		"key0": {
			"key1": {
				{
					{
						"key2": &color,
					},
				},
			},
		},
	})
	assert.Nil(t, rpcErr)
	assert.Equal(t, test.ColorRed, *result7["key0"]["key1"][0][0]["key2"])
}

type testService struct {
}

func (*testService) Increase(c *rpc.Context, request int32) (int32, *rpc.Error) {
	fmt.Printf("input value: %d\n", request)
	request++
	if request == 10 {
		return 0, &rpc.Error{
			Code:  code.Internal,
			Error: fmt.Errorf("special number found"),
		}
	}
	return request, nil
}

func (*testService) PrintObject(c *rpc.Context, request *test.Object) (*test.Object, *rpc.Error) {
	fmt.Printf("input object: %v\n", *request)
	timestamp := time.Unix(time.Now().Unix(), 0)
	color := test.ColorGreen
	return &test.Object{
		IntField:    456,
		TimeField:   &timestamp,
		StringField: "bar",
		EnumField:   &color,
		IntArray:    []int32{4, 5, 6},
		IntMap: map[string]int32{
			"4": 4,
			"5": 5,
			"6": 6,
		},
		ObjArray: []*sub.SubObject{{Value: 456}},
		ObjMap: map[string]*sub.SubObject{
			"bar": {Value: 456},
		},
	}, nil
}

func (*testService) PrintSubObject(c *rpc.Context, request *sub.SubObject) (*sub.SubObject, *rpc.Error) {
	fmt.Printf("input sub object: %v\n", *request)
	return &sub.SubObject{
		Value: 456,
	}, nil
}

func (*testService) PrintTime(c *rpc.Context, request *time.Time) (*time.Time, *rpc.Error) {
	fmt.Printf("input time: %v\n", *request)
	t := time.Now()
	return &t, nil
}

func (*testService) PrintArray(c *rpc.Context, request []int32) ([]int32, *rpc.Error) {
	fmt.Printf("input array: %v\n", request)
	return []int32{4, 5, 6}, nil
}

func (*testService) PrintEnum(c *rpc.Context, request *test.Color) (*test.Color, *rpc.Error) {
	reqColor, _ := request.ToString()
	fmt.Printf("input enum: %s\n", string(reqColor))
	color := test.ColorGreen
	return &color, nil
}

func (*testService) PrintMap(c *rpc.Context, request map[string]string) (map[string]string, *rpc.Error) {
	fmt.Printf("input map: %s\n", request)
	return map[string]string{
		"bar": "foo",
	}, nil
}

func (*testService) PrintNested(c *rpc.Context, request map[string]map[string][][]map[string]*test.Color) (map[string]map[string][][]map[string]*test.Color, *rpc.Error) {
	fmt.Printf("input nested color: %s\n", *request["key0"]["key1"][0][0]["key2"])
	*request["key0"]["key1"][0][0]["key2"] = test.ColorRed
	return request, nil
}
*/

type MiniObject struct {
	IntField int32 `disorder:"int_field" json:"int_field"`
}
