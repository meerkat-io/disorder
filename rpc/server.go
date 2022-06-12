package rpc

import (
	"bytes"
	"fmt"
	"net"

	"github.com/meerkat-lib/disorder"
	"github.com/meerkat-lib/disorder/rpc/code"
)

type Service interface {
	Handle(*Context, string, *disorder.Decoder) (interface{}, *Error)
}

type Server struct {
	socket   *net.TCPListener
	services map[string]Service
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]Service),
	}
}

func (s *Server) Listen(addr string) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	socket, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}
	s.socket = socket
	go s.listen()
	return nil
}

func (s *Server) Close() {
	s.socket.Close()
}

func (s *Server) RegisterService(serviceName string, service Service) {
	s.services[serviceName] = service
}

func (s *Server) listen() {
	for {
		if socket, err := s.socket.AcceptTCP(); err == nil {
			conn := newConnection(socket)
			go s.handle(conn)
		} else {
			return
		}
	}
}

func (s *Server) handle(c *connection) {
	defer c.close()
	data, err := c.receive()
	if err != nil {
		return
	}

	context := NewContext()
	reader := bytes.NewBuffer(data)
	d := disorder.NewDecoder(reader)
	err = d.Decode(context.Headers)
	if err != nil {
		s.writeError(c, code.InvalidRequest, err)
		return
	}
	var serviceName title
	err = d.Decode(&serviceName)
	if err != nil {
		s.writeError(c, code.InvalidRequest, err)
		return
	}
	var methodName title
	err = d.Decode(&methodName)
	if err != nil {
		s.writeError(c, code.InvalidRequest, err)
		return
	}
	service, exists := s.services[string(serviceName)]
	if !exists {
		s.writeError(c, code.Unavailable, fmt.Errorf("service \"%s\" not found", serviceName))
		return
	}

	response, status := service.Handle(context, string(methodName), d)
	if status != nil && status.Code != code.OK {
		s.writeError(c, status.Code, status.Error)
		return
	}
	writer := &bytes.Buffer{}
	e := disorder.NewEncoder(writer)
	_ = e.Encode(byte(code.OK))
	err = e.Encode(response)
	if err != nil {
		s.writeError(c, code.Internal, err)
		return
	}
	_ = c.send(writer.Bytes())
}

func (s *Server) writeError(c *connection, code code.Code, err error) {
	writer := &bytes.Buffer{}
	e := disorder.NewEncoder(writer)
	status := byte(code)
	_ = e.Encode(status)
	err = e.Encode(err.Error())
	if err != nil {
		return
	}
	_ = c.send(writer.Bytes())
}
