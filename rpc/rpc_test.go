package rpc_test

import (
	"fmt"
	"testing"

	"github.com/meerkat-io/disorder"
	"github.com/meerkat-io/disorder/rpc"
	"github.com/meerkat-io/disorder/rpc/code"
	"github.com/stretchr/testify/assert"
)

func TestRpcMath(t *testing.T) {
	s := rpc.NewServer()
	RegisterTestService(s, &testService{})
	err := s.Listen(":9999")
	assert.Nil(t, err)

	c := NewTestServiceClient("localhost:9999")
	result, rpcErr := c.Increase(int32(1))
	assert.Nil(t, rpcErr)
	assert.Equal(t, int32(2), result)

	if rpcErr != nil {
		fmt.Println(rpcErr.Error.Error())
	}
}

type TestService interface {
	Increase(int32) (int32, *rpc.Error)
}

type testServiceHandler struct {
	service TestService
}

func (h *testServiceHandler) Increase(d *disorder.Decoder) (interface{}, *rpc.Error) {
	var request int32
	err := d.Decode(&request)
	if err != nil {
		return nil, &rpc.Error{
			Code: code.InvalidRequest,
		}
	}
	return h.service.Increase(request)
}

type testService struct {
}

func (s *testService) Increase(request int32) (int32, *rpc.Error) {
	return request + 1, nil
}

func RegisterTestService(server *rpc.Server, service TestService) {
	handler := &testServiceHandler{service: service}
	server.RegisterHandler("test", "increase", handler.Increase)
}

type TestServiceClient struct {
	client *rpc.Client
}

func NewTestServiceClient(addr string) *TestServiceClient {
	return &TestServiceClient{client: rpc.NewClient(addr, "test")}
}

func NewTestServiceClientWithBalancer(b rpc.Balancer) *TestServiceClient {
	return &TestServiceClient{client: rpc.NewClientWithBalancer(b, "test")}
}

func (c *TestServiceClient) Increase(request int32) (int32, *rpc.Error) {
	var response int32
	err := c.client.Send("increase", request, &response)
	return response, err
}

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
