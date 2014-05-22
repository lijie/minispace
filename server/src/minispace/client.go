package minispace

import "code.google.com/p/go.net/websocket"

type Client struct {
	conn *websocket.Conn
	enable bool
}

func (c *Client) Proc() {
	return
}

func NewClient(conn *websocket.Conn) *Client {
	c := &Client{
		conn: conn,
		enable: true,
	}
	return c
}
