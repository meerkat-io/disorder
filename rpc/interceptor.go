package rpc

type ClientInterceptor interface {
	Intercept(context *Context) *Error
}

type ServerInterceptor interface {
	PreHandle(context *Context) *Error
	PostHandle(context *Context, err *Error)
}
