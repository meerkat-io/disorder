package rpc

import (
	"fmt"

	"github.com/meerkat-io/bloom/tcp"
	"github.com/meerkat-io/disorder"
	"github.com/meerkat-io/disorder/rpc/code"
)

type Handler func(decoder *disorder.Decoder) (interface{}, *Error)

type Server struct {
	listener     *tcp.Listener
	handlers     map[string]map[string]Handler
	interceptors map[string][]ServerInterceptor
}

func NewServer() *Server {
	return &Server{
		handlers:     make(map[string]map[string]Handler),
		interceptors: make(map[string][]ServerInterceptor),
	}
}

func (s *Server) Listen(addr string) error {
	l, err := tcp.Listen(addr, s)
	s.listener = l
	return err
}

func (s *Server) Close() {
	s.listener.Close()
}

func (s *Server) RegisterHandler(service, method string, handler Handler) {
	if _, ok := s.handlers[service]; !ok {
		s.handlers[service] = make(map[string]Handler)
	}
	s.handlers[service][method] = handler
}

func (s *Server) AddInterceptor(service string, interceptor ServerInterceptor) {
	s.interceptors[service] = append(s.interceptors[service], interceptor)
}

func (s *Server) Accept(conn *tcp.Connection) {
	defer conn.Close()

	// read
	d := disorder.NewDecoder(conn.Reader())
	context := NewContext()
	err := d.Decode(&context.headers)
	if err != nil {
		s.sendError(conn, code.InvalidRequest, err)
		return
	}
	service, method, err := context.readRpcInfo()
	if err != nil {
		s.sendError(conn, code.InvalidRequest, err)
		return
	}

	// handle
	_, exists := s.handlers[service]
	if !exists {
		s.sendError(conn, code.Unimplemented, fmt.Errorf("service \"%s\" not found", service))
		return
	}
	handler, exists := s.handlers[service][method]
	if !exists {
		s.sendError(conn, code.Unimplemented, fmt.Errorf("service \"%s\" does not contain method \"%s\"", service, method))
		return
	}
	var rpcErr *Error
	defer s.postHandle(service, context, rpcErr)
	rpcErr = s.preHandle(service, context)
	if rpcErr != nil {
		s.sendError(conn, rpcErr.Code, rpcErr.Error)
		return
	}
	response, rpcErr := handler(d)
	if rpcErr != nil {
		s.sendError(conn, rpcErr.Code, rpcErr.Error)
		return
	}

	// write
	rpcErr = s.sendResponse(conn, response)
}

func (s *Server) preHandle(service string, context *Context) *Error {
	if len(s.interceptors[service]) > 0 {
		for _, i := range s.interceptors[service] {
			rpcErr := i.PreHandle(context)
			if rpcErr != nil {
				return rpcErr
			}
		}
	}
	return nil
}

func (s *Server) postHandle(service string, context *Context, err *Error) {
	if len(s.interceptors[service]) > 0 {
		for _, i := range s.interceptors[service] {
			i.PostHandle(context, err)
		}
	}
}

func (s *Server) sendError(conn *tcp.Connection, code code.Code, err error) {
	context := NewContext()
	context.writeError(code, err)
	e := disorder.NewEncoder(conn.Writer())
	_ = e.Encode(context.headers)
}

func (s *Server) sendResponse(conn *tcp.Connection, response interface{}) *Error {
	e := disorder.NewEncoder(conn.Writer())
	err := e.Encode(map[string]string{})
	if err != nil {
		return &Error{
			Code:  code.Internal,
			Error: err,
		}
	}
	err = e.Encode(response)
	if err != nil {
		return &Error{
			Code:  code.Internal,
			Error: err,
		}
	}
	return nil
}
