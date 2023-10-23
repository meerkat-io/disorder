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

type dummyBalancer struct {
	addr string
}

func (b *dummyBalancer) Address() (string, error) {
	return b.addr, nil
}

type Client struct {
	b Balancer
}

func NewClient(addr string) *Client {
	return &Client{
		b: &dummyBalancer{
			addr: addr,
		},
	}
}

func NewClientWithBalancer(b Balancer) *Client {
	return &Client{
		b: b,
	}
}

func (c *Client) Send(context *Context, serviceName, methodName, entityName string, request interface{}, response interface{}) *Error {
	// dial
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

	// send
	e := disorder.NewEncoder(conn.Writer())
	if context == nil {
		context = NewContext()
	}
	err = context.writeRpcInfo(serviceName, methodName, entityName)
	if err != nil {
		return &Error{
			Code:  code.InvalidRequest,
			Error: err,
		}
	}
	err = e.Encode(context.headers)
	if err != nil {
		return &Error{
			Code:  code.InvalidRequest,
			Error: err,
		}
	}
	if entityName != "" {
		if request == nil {
			return &Error{
				Code:  code.InvalidRequest,
				Error: fmt.Errorf("request is null"),
			}
		}
		err = e.Encode(request)
		if err != nil {
			return &Error{
				Code:  code.InvalidRequest,
				Error: err,
			}
		}
	}

	// read
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
