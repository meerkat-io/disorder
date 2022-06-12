package rpc

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/meerkat-lib/disorder/rpc/code"
)

const (
	packetHeaderSize = 4
	maxPacketSize    = 1024 * 1024
)

type connection struct {
	socket *net.TCPConn
}

func newConnection(socket *net.TCPConn) *connection {
	return &connection{
		socket: socket,
	}
}

func (c *connection) receive() ([]byte, error) {
	var data []byte
	length, err := c.readPacketSize()
	if err == nil {
		readed := 0
		bytes := 0
		data = make([]byte, length)
		for readed < length && err == nil {
			bytes, err = c.socket.Read(data[readed:])
			readed += bytes
		}
	}
	return data, err
}

func (c *connection) send(data []byte) error {
	length := len(data)
	err := c.writePacketSize(length)
	if err == nil {
		writed := 0
		bytes := 0
		for writed < length && err == nil {
			bytes, err = c.socket.Write(data[writed:])
			writed += bytes
		}
	}
	return err
}

func (c *connection) close() {
	c.socket.Close()
}

func (c *connection) readPacketSize() (int, error) {
	header := make([]byte, packetHeaderSize)
	size, err := c.socket.Read(header)
	if err != nil {
		return 0, fmt.Errorf("read socket error from %s", c.socket.RemoteAddr())
	}
	if size != packetHeaderSize {
		return 0, fmt.Errorf("invalid header size")
	}
	length := int(binary.LittleEndian.Uint32(header))
	if length == 0 {
		return 0, fmt.Errorf("empty packet from %s", c.socket.RemoteAddr())
	}
	if length > maxPacketSize {
		return 0, fmt.Errorf("data overflow from %s", c.socket.RemoteAddr())
	}
	return length, nil
}

func (c *connection) writePacketSize(length int) error {
	if length > maxPacketSize {
		return fmt.Errorf("data overflow write to %s", c.socket.RemoteAddr())
	}
	header := make([]byte, packetHeaderSize)
	binary.LittleEndian.PutUint32(header, uint32(length))
	bytesWrite, err := c.socket.Write(header)
	if err != nil || bytesWrite != packetHeaderSize {
		return fmt.Errorf("write data error to %s", c.socket.RemoteAddr())
	}
	return nil
}

func (c *connection) writeError(code code.Code, err error) error {
	data := make([]byte, 5)
	data[0] = byte(code)
	message := err.Error()
	binary.LittleEndian.PutUint32(data[1:], uint32(len(message)))
	data = append(data, []byte(message)...)
	return c.send(data)
}

func (c *connection) writeCode(code code.Code) error {
	data := make([]byte, 1)
	data[0] = byte(code)
	_, err := c.socket.Write(data)
	return err
}
