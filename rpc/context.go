package rpc

import "github.com/meerkat-lib/disorder/rpc/code"

type Context struct {
	Headers map[string]string

	errorCode code.Code
	err       error
}

func NewContext() *Context {
	return &Context{
		Headers: make(map[string]string),
	}
}

func (c *Context) Error(code code.Code, err error) {
	c.errorCode = code
	c.err = err
	if err.Error() == "" {
		panic("empty error content")
	}
}
