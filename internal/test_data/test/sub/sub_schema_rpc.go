package sub

import (
	"fmt"

	"github.com/meerkat-lib/disorder"
	"github.com/meerkat-lib/disorder/rpc"
	"github.com/meerkat-lib/disorder/rpc/code"
)

type MathService interface {
	Increase(*rpc.Context, *int32) (*int32, *rpc.Error)
}

func NewMathServiceClient(client *rpc.Client) MathService {
	return &mathServiceClient{
		name:   "math_service",
		client: client,
	}
}

type mathServiceClient struct {
	name   string
	client *rpc.Client
}

func (c *mathServiceClient) Increase(context *rpc.Context, request *int32) (*int32, *rpc.Error) {
	var response int32
	err := c.client.Send(context, c.name, "increase", request, &response)
	return &response, err
}

type mathServiceServer struct {
	name    string
	service MathService
}

func RegisterMathService(s *rpc.Server, service MathService) {
	s.RegisterService("math_service", &mathServiceServer{
		name:    "math_service",
		service: service,
	})
}

func (s *mathServiceServer) Handle(context *rpc.Context, method string, d *disorder.Decoder) (interface{}, *rpc.Error) {
	switch {
	case method == "increase":
		var request int32
		err := d.Decode(&request)
		if err != nil {
			return nil, &rpc.Error{
				Code:  code.InvalidRequest,
				Error: err,
			}
		}
		response, rpcErr := s.service.Increase(context, &request)
		if rpcErr != nil {
			return nil, rpcErr
		}
		return response, nil

	default:
		return nil, &rpc.Error{
			Code:  code.Unimplemented,
			Error: fmt.Errorf("Unimplemented method \"%s\" under service \"%s\"", method, s.name),
		}
	}
}
