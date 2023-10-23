package rpc

import (
	"bytes"
	"fmt"

	"github.com/meerkat-io/bloom/tcp"

	"github.com/meerkat-io/disorder"
	"github.com/meerkat-io/disorder/rpc/code"
)

type Service interface {
	Handle(*Context, string, *disorder.Decoder) (interface{}, *Error)
}

type Server struct {
	listener *tcp.Listener
	services map[string]Service
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]Service),
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

func (s *Server) RegisterService(serviceName string, service Service) {
	s.services[serviceName] = service
}

func (s *Server) Accept(conn *tcp.Connection) {
	defer conn.Close()
	reader := conn.Reader()
	context := NewContext()
	d := disorder.NewDecoder(reader)
	err := d.Decode(context.headers)
	if err != nil {
		s.writeError(conn, context, code.InvalidRequest, err)
		return
	}
	serviceName, methodName, err := context.readRpcInfo()
	if err != nil {
		s.writeError(conn, context, code.InvalidRequest, err)
		return
	}
	service, exists := s.services[string(serviceName)]
	if !exists {
		s.writeError(conn, context, code.Unavailable, fmt.Errorf("service \"%s\" not found", serviceName))
		return
	}
	response, status := service.Handle(context, string(methodName), d)
	if status != nil {
		s.writeError(conn, context, status.Code, status.Error)
		return
	}

	writer := &bytes.Buffer{}
	_, _ = writer.Write([]byte{byte(code.OK)})
	e := disorder.NewEncoder(writer)
	err = e.Encode(response)
	if err != nil {
		s.writeError(conn, code.Internal, err)
		return
	}
	_ = conn.Send(writer.Bytes())
}

func (s *Server) writeError(conn *tcp.Connection, context *Context, code code.Code, err error) {
	context.writeError()
	writer := conn.Writer()
	e := disorder.NewEncoder(writer)
	status := byte(code)
	_ = e.Encode(status)
	err = e.Encode(err.Error())
	if err != nil {
		return
	}
	_ = conn.Send(writer.Bytes())
}
