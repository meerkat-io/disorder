package rpc

import (
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
	b            Balancer
	service      string
	interceptors []ClientInterceptor
}

func NewClient(addr, service string) *Client {
	return &Client{
		service: service,
		b: &dummyBalancer{
			addr: addr,
		},
	}
}

func NewClientWithBalancer(b Balancer, service string) *Client {
	return &Client{
		service: service,
		b:       b,
	}
}

func (c *Client) AddInterceptor(interceptor ClientInterceptor) {
	c.interceptors = append(c.interceptors, interceptor)
}

func (c *Client) Send(method string, request interface{}, response interface{}) *Error {
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

	// write
	e := disorder.NewEncoder(conn.Writer())
	context := NewContext()
	err = context.writeRpcInfo(c.service, method)
	if err != nil {
		return &Error{
			Code:  code.InvalidRequest,
			Error: err,
		}
	}
	rpcErr := c.intercept(context)
	if rpcErr != nil {
		return rpcErr
	}
	err = e.Encode(context.headers)
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

	// read
	d := disorder.NewDecoder(conn.Reader())
	context.headers = make(map[string]string)
	err = d.Decode(&context.headers)
	if err != nil {
		return &Error{
			Code:  code.DataCorrupt,
			Error: err,
		}
	}
	rpcErr = context.readError()
	if rpcErr != nil {
		return rpcErr
	}
	err = d.Decode(response)
	if err != nil {
		return &Error{
			Code:  code.DataCorrupt,
			Error: err,
		}
	}
	return nil
}

func (c *Client) intercept(context *Context) *Error {
	if len(c.interceptors) > 0 {
		for _, i := range c.interceptors {
			rpcErr := i.Intercept(context)
			if rpcErr != nil {
				return rpcErr
			}
		}
	}
	return nil
}
