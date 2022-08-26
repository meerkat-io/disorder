// Code generated by https://github.com/meerkat-io/disorder; DO NOT EDIT.
package test

import (
	"fmt"
	"github.com/meerkat-io/disorder"
	"github.com/meerkat-io/disorder/internal/test_data/test/sub"
	"github.com/meerkat-io/disorder/rpc"
	"github.com/meerkat-io/disorder/rpc/code"
	"time"
)

type primaryServiceHandler func(*rpc.Context, *disorder.Decoder) (interface{}, *rpc.Error)

type PrimaryService interface {
	PrintMap(*rpc.Context, map[string]string) (map[string]string, *rpc.Error)
	PrintObject(*rpc.Context, *Object) (*Object, *rpc.Error)
	PrintSubObject(*rpc.Context, *sub.SubObject) (*sub.SubObject, *rpc.Error)
	PrintTime(*rpc.Context, *time.Time) (*time.Time, *rpc.Error)
	PrintArray(*rpc.Context, []int32) ([]int32, *rpc.Error)
	PrintEnum(*rpc.Context, *Color) (*Color, *rpc.Error)
}

func NewPrimaryServiceClient(client *rpc.Client) PrimaryService {
	return &primaryServiceClient{
		name:   "primary_service",
		client: client,
	}
}

type primaryServiceClient struct {
	name   string
	client *rpc.Client
}

func (c *primaryServiceClient) PrintMap(context *rpc.Context, request map[string]string) (map[string]string, *rpc.Error) {
	var response map[string]string = make(map[string]string)
	err := c.client.Send(context, c.name, "print_map", request, &response)
	return response, err
}

func (c *primaryServiceClient) PrintObject(context *rpc.Context, request *Object) (*Object, *rpc.Error) {
	var response *Object = &Object{}
	err := c.client.Send(context, c.name, "print_object", request, response)
	return response, err
}

func (c *primaryServiceClient) PrintSubObject(context *rpc.Context, request *sub.SubObject) (*sub.SubObject, *rpc.Error) {
	var response *sub.SubObject = &sub.SubObject{}
	err := c.client.Send(context, c.name, "print_sub_object", request, response)
	return response, err
}

func (c *primaryServiceClient) PrintTime(context *rpc.Context, request *time.Time) (*time.Time, *rpc.Error) {
	var response *time.Time = &time.Time{}
	err := c.client.Send(context, c.name, "print_time", request, response)
	return response, err
}

func (c *primaryServiceClient) PrintArray(context *rpc.Context, request []int32) ([]int32, *rpc.Error) {
	var response []int32
	err := c.client.Send(context, c.name, "print_array", request, &response)
	return response, err
}

func (c *primaryServiceClient) PrintEnum(context *rpc.Context, request *Color) (*Color, *rpc.Error) {
	var response *Color = new(Color)
	err := c.client.Send(context, c.name, "print_enum", request, response)
	return response, err
}

type primaryServiceServer struct {
	name    string
	service PrimaryService
	methods map[string]primaryServiceHandler
}

func RegisterPrimaryServiceServer(s *rpc.Server, service PrimaryService) {
	server := &primaryServiceServer{
		name:    "primary_service",
		service: service,
	}
	server.methods = map[string]primaryServiceHandler{
		"print_map":        server.printMap,
		"print_object":     server.printObject,
		"print_sub_object": server.printSubObject,
		"print_time":       server.printTime,
		"print_array":      server.printArray,
		"print_enum":       server.printEnum,
	}
	s.RegisterService("primary_service", server)
}

func (s *primaryServiceServer) Handle(context *rpc.Context, method string, d *disorder.Decoder) (interface{}, *rpc.Error) {
	handler, ok := s.methods[method]
	if ok {
		return handler(context, d)
	}
	return nil, &rpc.Error{
		Code:  code.Unimplemented,
		Error: fmt.Errorf("Unimplemented method \"%s\" under service \"%s\"", method, s.name),
	}
}

func (s *primaryServiceServer) printMap(context *rpc.Context, d *disorder.Decoder) (interface{}, *rpc.Error) {
	var request map[string]string = make(map[string]string)
	err := d.Decode(&request)
	if err != nil {
		return nil, &rpc.Error{
			Code:  code.InvalidRequest,
			Error: err,
		}
	}
	response, rpcErr := s.service.PrintMap(context, request)
	if rpcErr != nil {
		return nil, rpcErr
	}
	return response, nil
}

func (s *primaryServiceServer) printObject(context *rpc.Context, d *disorder.Decoder) (interface{}, *rpc.Error) {
	var request *Object = &Object{}
	err := d.Decode(request)
	if err != nil {
		return nil, &rpc.Error{
			Code:  code.InvalidRequest,
			Error: err,
		}
	}
	response, rpcErr := s.service.PrintObject(context, request)
	if rpcErr != nil {
		return nil, rpcErr
	}
	return response, nil
}

func (s *primaryServiceServer) printSubObject(context *rpc.Context, d *disorder.Decoder) (interface{}, *rpc.Error) {
	var request *sub.SubObject = &sub.SubObject{}
	err := d.Decode(request)
	if err != nil {
		return nil, &rpc.Error{
			Code:  code.InvalidRequest,
			Error: err,
		}
	}
	response, rpcErr := s.service.PrintSubObject(context, request)
	if rpcErr != nil {
		return nil, rpcErr
	}
	return response, nil
}

func (s *primaryServiceServer) printTime(context *rpc.Context, d *disorder.Decoder) (interface{}, *rpc.Error) {
	var request *time.Time = &time.Time{}
	err := d.Decode(request)
	if err != nil {
		return nil, &rpc.Error{
			Code:  code.InvalidRequest,
			Error: err,
		}
	}
	response, rpcErr := s.service.PrintTime(context, request)
	if rpcErr != nil {
		return nil, rpcErr
	}
	return response, nil
}

func (s *primaryServiceServer) printArray(context *rpc.Context, d *disorder.Decoder) (interface{}, *rpc.Error) {
	var request []int32
	err := d.Decode(&request)
	if err != nil {
		return nil, &rpc.Error{
			Code:  code.InvalidRequest,
			Error: err,
		}
	}
	response, rpcErr := s.service.PrintArray(context, request)
	if rpcErr != nil {
		return nil, rpcErr
	}
	return response, nil
}

func (s *primaryServiceServer) printEnum(context *rpc.Context, d *disorder.Decoder) (interface{}, *rpc.Error) {
	var request *Color = new(Color)
	err := d.Decode(request)
	if err != nil {
		return nil, &rpc.Error{
			Code:  code.InvalidRequest,
			Error: err,
		}
	}
	response, rpcErr := s.service.PrintEnum(context, request)
	if rpcErr != nil {
		return nil, rpcErr
	}
	return response, nil
}
