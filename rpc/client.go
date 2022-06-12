package rpc

/*
import (
	"net"
	"time"

	"github.com/meerkat-lib/house/order"
)

const timeout time.Duration = time.Second * 5

type Client struct {
	addr string
}

func dial(addr string) (*connection, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	return &connection{
		socket: conn,
	}, nil
}

func NewClient(addr string) (*Client, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	return &Client{
		addr: addr,
	}
}

// Call call remote method
func (c *Client) Call(req order.BinaryMarshaller, res order.BinaryMarshaller) (err error) {
	var conn *Connection
	if conn, err = Dial(c.addr, c.bufferSize); err == nil {
		if err = conn.SendObject(req, c.auth); err == nil {
			err = conn.ReceiveObject(res)
		}
		conn.Close()
	}
	return
}

// CallRaw call remote method by bytes
func (c *Client) CallRaw(req []byte) (res []byte, err error) {
	var conn *Connection
	if conn, err = Dial(c.addr, c.bufferSize); err == nil {
		if err = conn.Send(req); err == nil {
			res, err = conn.Receive()
		}
		conn.Close()
	}
	return
}

// SetAuth set auth info
func (c *Client) SetAuth(auth *Token) {
	c.auth = auth
}
*/
