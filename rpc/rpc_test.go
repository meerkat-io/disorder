package rpc_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/meerkat-io/disorder"
	"github.com/meerkat-io/disorder/rpc"
	"github.com/meerkat-io/disorder/rpc/code"
	"github.com/stretchr/testify/assert"
)

func TestRpc(t *testing.T) {
	s := rpc.NewServer()
	RegisterTestService(s, &TestServiceImpl{})
	err := s.Listen(":9999")
	assert.Nil(t, err)

	c := NewTestServiceClient("localhost:9999")

	increase, rpcErr := c.Increase(int32(1))
	assert.Nil(t, rpcErr)
	assert.Equal(t, int32(2), increase)

	_, rpcErr = c.Decrease(int32(9))
	assert.NotNil(t, rpcErr)
	assert.Equal(t, code.Unimplemented, rpcErr.Code)

	timestamp := time.UnixMilli(time.Now().UnixMilli())
	request := &Object{
		IntField:    123,
		TimeField:   &timestamp,
		StringField: "foo",
		EnumField:   &ColorBlue,
		IntArray:    []int32{1, 2, 3},
		IntMap: map[string]int32{
			"1": 1,
			"2": 2,
			"3": 3,
		},
		ObjArray: []*NumberWrapper{{Value: &Number{Value: 123}}},
		ObjMap: map[string]*NumberWrapper{
			"foo": {Value: &Number{Value: 123}},
		},
		Nested: map[string]map[string][][]map[string]*Color{
			"key0": {
				"key1": {
					{
						{
							"key2": &ColorRed,
						},
					},
				},
			},
		},
	}
	response, rpcErr := c.Reflect(request)
	assert.Nil(t, rpcErr)
	assert.Equal(t, request, response)

	s.Close()
}

func TestInterceptor(t *testing.T) {
	s := rpc.NewServer()
	RegisterTestService(s, &TestServiceImpl{})
	metrics := &ServerMetricsInterceptor{}
	AddTestServiceInterceptor(s, metrics)
	AddTestServiceInterceptor(s, &ServerAuthInterceptor{})
	err := s.Listen(":9999")
	assert.Nil(t, err)

	c := NewTestServiceClient("localhost:9999")

	_, rpcErr := c.Increase(int32(1))
	assert.NotNil(t, rpcErr)
	assert.Equal(t, rpcErr.Code, code.Unauthenticated)

	c.AddInterceptor(&ClientAuthInterceptor{})

	increase, rpcErr := c.Increase(int32(1))
	assert.Nil(t, rpcErr)
	assert.Equal(t, int32(2), increase)

	assert.Equal(t, 2, metrics.counter)
	assert.Equal(t, 1, metrics.errCounter)
	assert.True(t, metrics.delay > 0)

	s.Close()
}

type TestService interface {
	Increase(int32) (int32, *rpc.Error)
	Relect(*Object) (*Object, *rpc.Error)
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

func (h *testServiceHandler) Relect(d *disorder.Decoder) (interface{}, *rpc.Error) {
	var request Object
	err := d.Decode(&request)
	if err != nil {
		return nil, &rpc.Error{
			Code: code.InvalidRequest,
		}
	}
	return h.service.Relect(&request)
}

type TestServiceImpl struct {
}

func (s *TestServiceImpl) Increase(request int32) (int32, *rpc.Error) {
	return request + 1, nil
}

func (s *TestServiceImpl) Relect(request *Object) (*Object, *rpc.Error) {
	return request, nil
}

func RegisterTestService(server *rpc.Server, service TestService) {
	handler := &testServiceHandler{service: service}
	server.RegisterHandler("test", "increase", handler.Increase)
	server.RegisterHandler("test", "reflect", handler.Relect)
}

func AddTestServiceInterceptor(server *rpc.Server, interceptor rpc.ServerInterceptor) {
	server.AddInterceptor("test", interceptor)
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

func (c *TestServiceClient) AddInterceptor(interceptor rpc.ClientInterceptor) {
	c.client.AddInterceptor(interceptor)
}

func (c *TestServiceClient) Increase(request int32) (int32, *rpc.Error) {
	var response int32
	err := c.client.Send("increase", request, &response)
	return response, err
}

func (c *TestServiceClient) Decrease(request int32) (int32, *rpc.Error) {
	var response int32
	err := c.client.Send("decrease", request, &response)
	return response, err
}

func (c *TestServiceClient) Reflect(request *Object) (*Object, *rpc.Error) {
	var response Object
	err := c.client.Send("reflect", request, &response)
	return &response, err
}

type ClientAuthInterceptor struct {
}

func (c *ClientAuthInterceptor) Intercept(context *rpc.Context) *rpc.Error {
	context.SetHeader("user", "user")
	context.SetHeader("password", "password")
	return nil
}

type ServerAuthInterceptor struct {
}

func (s *ServerAuthInterceptor) PreHandle(context *rpc.Context) *rpc.Error {
	if context.GetHeader("user") == "user" && context.GetHeader("password") == "password" {
		return nil
	}
	return &rpc.Error{
		Code:  code.Unauthenticated,
		Error: fmt.Errorf("unauthenticated session"),
	}
}

func (s *ServerAuthInterceptor) PostHandle(context *rpc.Context, err *rpc.Error) {
}

type ServerMetricsInterceptor struct {
	counter    int
	errCounter int
	delay      int64
}

func (s *ServerMetricsInterceptor) PreHandle(context *rpc.Context) *rpc.Error {
	s.counter++
	context.SetHeader("start", strconv.FormatInt(time.Now().UnixNano(), 10))
	return nil
}

func (s *ServerMetricsInterceptor) PostHandle(context *rpc.Context, err *rpc.Error) {
	fmt.Println("post handle", err)
	if err == nil {
		s.delay, _ = strconv.ParseInt(context.GetHeader("start"), 10, 64)
	} else {
		s.errCounter++
	}
}
