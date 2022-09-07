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
	data, err := conn.Receive()
	if err != nil {
		return
	}
	context := NewContext()
	reader := bytes.NewBuffer(data)
	d := disorder.NewDecoder(reader)
	err = d.Decode(context.Headers)
	if err != nil {
		s.writeError(conn, code.InvalidRequest, err)
		return
	}
	var serviceName title
	err = d.Decode(&serviceName)
	if err != nil {
		s.writeError(conn, code.InvalidRequest, err)
		return
	}
	var methodName title
	err = d.Decode(&methodName)
	if err != nil {
		s.writeError(conn, code.InvalidRequest, err)
		return
	}
	service, exists := s.services[string(serviceName)]
	if !exists {
		s.writeError(conn, code.Unavailable, fmt.Errorf("service \"%s\" not found", serviceName))
		return
	}
	response, status := service.Handle(context, string(methodName), d)
	if status != nil && status.Code != code.OK {
		s.writeError(conn, status.Code, status.Error)
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

func (s *Server) writeError(conn *tcp.Connection, code code.Code, err error) {
	writer := &bytes.Buffer{}
	e := disorder.NewEncoder(writer)
	status := byte(code)
	_ = e.Encode(status)
	err = e.Encode(err.Error())
	if err != nil {
		return
	}
	_ = conn.Send(writer.Bytes())
}
