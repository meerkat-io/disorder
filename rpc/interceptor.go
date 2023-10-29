package rpc

type Interceptor interface {
	Intercept(context *Context) *Error
}
