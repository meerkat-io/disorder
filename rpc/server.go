package rpc

import (
	"bytes"
	"fmt"
	"net"

	"github.com/meerkat-lib/disorder"
	"github.com/meerkat-lib/disorder/internal/utils/logger"
	"github.com/meerkat-lib/disorder/rpc/code"
)

type Service interface {
	Handle(*Context, *disorder.Decoder) interface{}
}

type Server struct {
	socket   *net.TCPListener
	services map[string]Service
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Listen(addr string) error {
	logger.Infof("listening %s", addr)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		logger.Errorf("listen %s failed: %s", addr, err.Error())
		return err
	}
	socket, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		logger.Errorf("listen %s failed: %s", addr, err.Error())
		return err
	}
	s.socket = socket
	go s.listen()
	logger.Infof("started listen on %s ", socket.Addr())
	return nil
}

func (s *Server) Close() {
	logger.Infof("stopped listen on %s ", s.socket.Addr())
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
		_ = c.writeError(code.InvalidRequest, err)
		return
	}

	serviceName := ""
	err = d.Decode(serviceName)
	if err != nil {
		_ = c.writeError(code.InvalidRequest, err)
		return
	}

	service, exists := s.services[serviceName]
	if !exists {
		_ = c.writeError(code.Unavailable, fmt.Errorf("service \"%s\" not found", serviceName))
		return
	}

	response := service.Handle(context, d)
	if context.errorCode != code.OK {
		_ = c.writeError(context.errorCode, context.err)
		return
	}

	writer := &bytes.Buffer{}
	e := disorder.NewEncoder(writer)
	err = e.Encode(response)
	if err != nil {
		_ = c.writeError(code.Internal, err)
		return
	}
	err = c.writeCode(code.OK)
	if err == nil {
		_ = c.send(writer.Bytes())
	}
}
