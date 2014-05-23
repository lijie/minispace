// Copyright (c) LiJie 2014-05-23 17:32:46

package minispace

import "code.google.com/p/go.net/websocket"
import "time"
import "errors"
import "fmt"

type Msg struct {
	Cmd float64
	ErrCode float64
	Seq float64
	Userid string
	Body map[string]interface{}
}

type Client struct {
	User
	conn *websocket.Conn
	enable bool
	msg Msg
}

type ClientProc func(*Client, *Msg) int

var procFuncArray [128]ClientProc
// Register your client cmd proc function
func ClientProcRegister(cmd int, proc ClientProc) error {
	if cmd >= 128 {
		return errors.New("Command is too big")
	}

	if procFuncArray[cmd] != nil {
		return errors.New("Command is already registed")
	}

	procFuncArray[cmd] = proc
	return nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) readPacket(conn *websocket.Conn) (*Msg, error) {
	var err error
	if err = websocket.JSON.Receive(conn, &c.msg); err != nil {
		return nil, err
	} else {
		fmt.Println(c.msg)
		return &c.msg, nil
	}
}

func (c *Client) procMsg(msg *Msg) {
}

func (c *Client) procTimeout() {
	c.enable = false
}

func (c *Client) readMsg(req_chan chan *Msg) {
	for {
		if msg, err := c.readPacket(c.conn); err != nil {
			c.enable = false
			req_chan <- nil
			return
		} else {
			req_chan <- msg
		}
	}
}

func (c *Client) Proc() {
	defer c.Close()

	req_chan := make(chan *Msg, 1)
	go c.readMsg(req_chan)

	for c.enable {
		select {
		case msg := <-req_chan:
			c.procMsg(msg)
		case <-time.After(300 * time.Second):
			c.procTimeout()
		}
	}
	return
}

func NewClient(conn *websocket.Conn) *Client {
	c := &Client{
		conn: conn,
		enable: true,
	}
	return c
}
