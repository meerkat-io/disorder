package rpc

import (
	"fmt"
	"strconv"

	"github.com/meerkat-io/disorder/rpc/code"
)

const (
	serviceName = "service"
	methodName  = "method"
	errorCode   = "code"
	errorMsg    = "error"
)

var reservedHeader = map[string]bool{
	serviceName: true,
	methodName:  true,
	errorCode:   true,
	errorMsg:    true,
}

type Context struct {
	headers map[string]string
}

func NewContext() *Context {
	return &Context{
		headers: make(map[string]string),
	}
}

func (c *Context) GetHeader(key string) (string, error) {
	if reservedHeader[key] {
		return "", fmt.Errorf("\"%s\" is reserved", key)
	}
	return c.headers[key], nil
}

func (c *Context) SetHeader(key, value string) error {
	if reservedHeader[key] {
		return fmt.Errorf("\"%s\" is reserved", key)
	}
	c.headers[key] = value
	return nil
}

func (c *Context) UnsetHeader(key string) error {
	if reservedHeader[key] {
		return fmt.Errorf("\"%s\" is reserved", key)
	}
	delete(c.headers, key)
	return nil
}

func (c *Context) Router() string {
	service, method, err := c.readRpcInfo()
	if err != nil {
		return "unknown"
	}
	return fmt.Sprintf("%s.%s", service, method)
}

func (c *Context) readRpcInfo() (service, method string, err error) {
	if c.headers[serviceName] == "" || c.headers[methodName] == "" {
		err = fmt.Errorf("invalid rpc info")
		return
	}
	service = c.headers[serviceName]
	method = c.headers[methodName]
	return
}

func (c *Context) writeRpcInfo(service, method string) error {
	if service == "" || method == "" {
		return fmt.Errorf("invalid rpc info")
	}
	c.headers[serviceName] = service
	c.headers[methodName] = method
	return nil
}

func (c *Context) readError() *Error {
	if c.headers[errorCode] == "" && c.headers[errorMsg] == "" {
		return nil
	}
	i, err := strconv.Atoi(c.headers[errorCode])
	if err != nil {
		return &Error{
			Code:  code.Unknown,
			Error: err,
		}
	}
	return &Error{
		Code:  code.Code(i),
		Error: fmt.Errorf("%s", c.headers[errorMsg]),
	}
}

func (c *Context) writeError(code code.Code, err error) {
	c.headers[errorCode] = strconv.Itoa(int(code))
	c.headers[errorMsg] = err.Error()
}
