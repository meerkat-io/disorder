package rpc

import (
	"bytes"
	"fmt"
	"net"

	"github.com/meerkat-lib/disorder"
	"github.com/meerkat-lib/disorder/rpc/code"
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

type title string

func (t *title) Enum() {}
func (t *title) FromString(enum string) error {
	*t = title(enum)
	return nil
}
func (t *title) ToString() (string, error) {
	return string(*t), nil
}

type Client struct {
	b Balancer
}

func Dial(addr string) *Client {
	return &Client{
		b: &balancerImpl{
			addr: addr,
		},
	}
}

func WithBalancer(b Balancer) *Client {
	return &Client{
		b: b,
	}
}

func (c *Client) dial() (*connection, error) {
	addr, err := c.b.Address()
	if err != nil {
		return nil, err
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	return newConnection(conn), nil
}

func (c *Client) Send(context *Context, serviceName, methodName string, request interface{}, response interface{}) *Error {
	conn, err := c.dial()
	if err == nil {
		defer conn.close()

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
		var service title = title(serviceName)
		err = e.Encode(service)
		if err != nil {
			return &Error{
				Code:  code.InvalidRequest,
				Error: err,
			}
		}
		var method title = title(methodName)
		err = e.Encode(method)
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
		err = conn.send(writer.Bytes())
		if err != nil {
			return &Error{
				Code:  code.NetworkDisconnected,
				Error: err,
			}
		}

		var data []byte
		data, err = conn.receive()
		if err != nil {
			return &Error{
				Code:  code.NetworkDisconnected,
				Error: err,
			}
		}
		status := code.Code(data[0])
		if status != code.OK {
			var errMsg string
			err = disorder.Unmarshal(data[1:], &errMsg)
			if err != nil {
				return &Error{
					Code:  code.Internal,
					Error: err,
				}
			}
			return &Error{
				Code:  status,
				Error: fmt.Errorf(errMsg),
			}
		}
		err = disorder.Unmarshal(data[1:], response)
	}
	if err != nil {
		return &Error{
			Code:  code.Internal,
			Error: err,
		}
	}
	return nil
}
