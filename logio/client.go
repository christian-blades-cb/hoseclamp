package logio

import (
	"fmt"
	"net"
)

type Client struct {
	connection *net.Conn
}

func NewClient(server string) (*Client, error) {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		return nil, err
	}

	return &Client{connection: &conn}, nil
}

func (c *Client) Log(line *LogLine) {
	fmt.Fprint(*c.connection, line.Serialize())
}

func (c *Client) Close() {
	conn := *c.connection
	conn.Close()
}

type LogLine struct {
	Node    string
	Stream  string
	Level   string
	Message string
}

func (ll *LogLine) Serialize() string {
	return fmt.Sprintf("+log|%s|%s|%s|%s\r\n", ll.Node, ll.Stream, ll.Level, ll.Message)
}
