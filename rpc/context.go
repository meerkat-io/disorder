package rpc

import "github.com/meerkat-io/disorder/rpc/code"

type Context struct {
	Headers map[string]string
}

func NewContext() *Context {
	return &Context{
		Headers: make(map[string]string),
	}
}

func (c *Context) AddHeader(key, value string) {
	c.Headers[key] = value
}

func (c *Context) RemoveHeader(key string) {
	delete(c.Headers, key)
}

type Error struct {
	Code  code.Code
	Error error
}
