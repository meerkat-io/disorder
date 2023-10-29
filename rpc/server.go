package rpc

import (
	"fmt"
	"reflect"

	"github.com/meerkat-io/bloom/tcp"
	"github.com/meerkat-io/disorder"
	"github.com/meerkat-io/disorder/rpc/code"
)

type Handler func(request interface{}) (interface{}, *Error)

type Server struct {
	listener     *tcp.Listener
	requests     map[string]map[string]reflect.Type
	handlers     map[string]map[string]Handler
	interceptors map[string][]Interceptor
}

func NewServer() *Server {
	return &Server{
		handlers:     make(map[string]map[string]Handler),
		requests:     make(map[string]map[string]reflect.Type),
		interceptors: make(map[string][]Interceptor),
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

func (s *Server) RegisterHandler(service, method string, handler Handler, request interface{}) {
	if _, ok := s.handlers[service]; !ok {
		s.handlers[service] = make(map[string]Handler)
		s.requests[service] = make(map[string]reflect.Type)
	}
	s.handlers[service][method] = handler
	s.requests[service][method] = reflect.TypeOf(request)
}

func (s *Server) AddInterceptor(service string, interceptor Interceptor) {
	s.interceptors[service] = append(s.interceptors[service], interceptor)
}

func (s *Server) Accept(conn *tcp.Connection) {
	defer conn.Close()
	reader := conn.Reader()
	context := NewContext()

	// read
	d := disorder.NewDecoder(reader)
	err := d.Decode(&context.headers)
	if err != nil {
		s.sendError(conn, context, code.InvalidRequest, err)
		return
	}
	service, method, err := context.readRpcInfo()
	if err != nil {
		s.sendError(conn, context, code.InvalidRequest, err)
		return
	}
	_, exists := s.handlers[service]
	if !exists {
		s.sendError(conn, context, code.Unimplemented, fmt.Errorf("service \"%s\" not found", service))
		return
	}
	handler, exists := s.handlers[service][method]
	if !exists {
		s.sendError(conn, context, code.Unimplemented, fmt.Errorf("service \"%s\" does not contain method \"%s\"", service, method))
		return
	}
	if len(s.interceptors[service]) > 0 {
		for _, i := range s.interceptors[service] {
			rpcErr := i.Intercept(context)
			if rpcErr != nil {
				s.sendError(conn, context, rpcErr.Code, rpcErr.Error)
				return
			}
		}
	}
	request := reflect.New(s.requests[service][method]).Elem().Interface()
	err = d.Decode(&request)
	if err != nil {
		s.sendError(conn, context, code.InvalidRequest, err)
		return
	}
	response, rpcErr := handler(&request)
	if rpcErr != nil {
		s.sendError(conn, context, rpcErr.Code, rpcErr.Error)
		return
	}

	// write
	e := disorder.NewEncoder(conn.Writer())
	_ = e.Encode(response)
}

func (s *Server) sendError(conn *tcp.Connection, context *Context, code code.Code, err error) {
	context.writeError(code, err)
	e := disorder.NewEncoder(conn.Writer())
	_ = e.Encode(context)
}
