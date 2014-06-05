// Copyright (c) LiJie 2014-05-23 17:32:46

package minispace

import "code.google.com/p/go.net/websocket"
import _ "time"
import "errors"
import "fmt"

type Msg struct {
	Cmd float64 `json:"cmd"`
	ErrCode float64 `json:"errcode"`
	Seq float64 `json:"seq"`
	Userid string `json:"userid"`
	Body map[string]interface{} `json:"body"`
}

type Packet struct {
	msg Msg
	client *Client
	ok bool
}

func NewMsg() *Msg {
	msg := &Msg{
		Body: make(map[string]interface{}),
	}
	return msg
}

func NewPacket(c *Client) *Packet {
	p := &Packet{
		client: c,
	}
	p.msg.Body = make(map[string]interface{})
	return p
}

type Client struct {
	User
	conn *websocket.Conn
	enable bool
	login bool
	scene *Scene
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

func (c *Client) readPacket(conn *websocket.Conn) (*Packet, error) {
	var err error
	p := NewPacket(c)

	p.ok = true
	if err = websocket.JSON.Receive(conn, &p.msg); err != nil {
		p.ok = false
		return p, err
	} else {
		return p, nil
	}
}

func (c *Client) Reply(msg *Msg) {
//	fmt.Printf("reply %v\n", msg)
	websocket.JSON.Send(c.conn, msg)
}

// called in scene goroutine
func (c *Client) ProcMsg(msg *Msg) {
//	fmt.Println(*msg)
	proc := procFuncArray[int(msg.Cmd)]
	if proc != nil {
//		fmt.Printf("proc %v\n", msg)
		proc(c, msg)
	}
}

func (c *Client) procTimeout() {
	c.enable = false
}

func (c *Client) Proc() {
	defer c.Close()
	ch, err := CurrentScene().AddClient(c)
	if err != nil {
		fmt.Printf("add client err")
		return
	}

	var p *Packet

	for {
		p, err = c.readPacket(c.conn)
		// send packet to current scene
		ch <- p

		if err != nil || !p.ok {
			// error, close conn
			c.enable = false
			break
		}
	}

	return
}

func NewClient(conn *websocket.Conn) *Client {
	c := &Client{
		conn: conn,
		enable: true,
		login: false,
	}
	InitUser(&c.User)
	return c
}
