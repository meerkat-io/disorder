package rpc

import (
	"bytes"
	"fmt"

	"github.com/meerkat-io/bloom/tcp"

	"github.com/meerkat-io/disorder"
	"github.com/meerkat-io/disorder/rpc/code"
)

type Balancer interface {
	Address() (string, error)
}

type balancerImpl struct {
	addr string
}

func (b *balancerImpl) Address() (string, error) {
	return b.addr, nil
}

type Client struct {
	b Balancer
}

func NewClient(addr string) *Client {
	return &Client{
		b: &balancerImpl{
			addr: addr,
		},
	}
}

func NewClientWithBalancer(b Balancer) *Client {
	return &Client{
		b: b,
	}
}

func (c *Client) Send(context *Context, serviceName, methodName string, request interface{}, response interface{}) *Error {
	addr, err := c.b.Address()
	if err != nil {
		return &Error{
			Code:  code.InvalidHost,
			Error: err,
		}
	}
	conn, err := tcp.Dial(addr)
	if err != nil {
		return &Error{
			Code:  code.NetworkDisconnected,
			Error: err,
		}
	}
	defer conn.Close()
	writer := &bytes.Buffer{}
	e := disorder.NewEncoder(writer)
	if context == nil {
		context = NewContext()
	}
	err = e.Encode(context.Headers)
	if err != nil {
		return &Error{
			Code:  code.InvalidRequest,
			Error: err,
		}
	}
	service := title(serviceName)
	err = e.Encode(&service)
	if err != nil {
		return &Error{
			Code:  code.InvalidRequest,
			Error: err,
		}
	}
	method := title(methodName)
	err = e.Encode(&method)
	if err != nil {
		return &Error{
			Code:  code.InvalidRequest,
			Error: err,
		}
	}
	err = e.Encode(request)
	if err != nil {
		return &Error{
			Code:  code.InvalidRequest,
			Error: err,
		}
	}
	err = conn.Send(writer.Bytes())
	if err != nil {
		return &Error{
			Code:  code.NetworkDisconnected,
			Error: err,
		}
	}
	var data []byte
	data, err = conn.Receive()
	if err != nil {
		return &Error{
			Code:  code.NetworkDisconnected,
			Error: err,
		}
	}
	status := data[0]
	d := disorder.NewDecoder(bytes.NewBuffer(data[1:]))
	if code.Code(status) != code.OK {
		var errMsg string
		err = d.Decode(&errMsg)
		if err != nil {
			return &Error{
				Code:  code.Internal,
				Error: err,
			}
		}
		return &Error{
			Code:  code.Code(status),
			Error: fmt.Errorf(errMsg),
		}
	}
	err = d.Decode(response)
	if err != nil {
		return &Error{
			Code:  code.Internal,
			Error: err,
		}
	}
	return nil
}
