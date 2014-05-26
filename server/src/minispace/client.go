// Copyright (c) LiJie 2014-05-23 17:32:46

package minispace

import "code.google.com/p/go.net/websocket"
import _ "time"
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
	scene *Scene
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
//		fmt.Println(c.msg)
		return &c.msg, nil
	}
}

func (c *Client) Reply(msg *Msg) {
	fmt.Printf("reply %v\n", msg)
	websocket.JSON.Send(c.conn, msg)
}

func (c *Client) ProcMsg(msg *Msg) {
//	fmt.Println(*msg)
	proc := procFuncArray[int(msg.Cmd)]
	if proc != nil {
		proc(c, msg)
	}
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
	ch, err := CurrentScene().AddClient(c)
	if err != nil {
		fmt.Printf("add client err")
		return
	}

	for {
		if _, err := c.readPacket(c.conn); err != nil {
			c.enable = false
			ch <- nil
		} else {
			ch <- c
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
