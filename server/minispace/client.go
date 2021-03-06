// Copyright (c) LiJie 2014-05-23 17:32:46

package minispace

import (
	"code.google.com/p/go.net/websocket"
	"errors"
	"fmt"
	"syscall"
	_ "time"
)

const (
	PROC_OK   = 0
	PROC_ERR  = 1
	PROC_KICK = 2
)

// Msg from client
type Msg struct {
	Cmd     float64                `json:"cmd"`
	ErrCode float64                `json:"errcode"`
	Seq     float64                `json:"seq"`
	Userid  string                 `json:"userid"`
	Body    map[string]interface{} `json:"body"`
}

// Msg wrapper for scene
type Packet struct {
	Msg
	client *Client
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
	p.Body = make(map[string]interface{})
	return p
}

type Client struct {
	User
	conn *websocket.Conn
}

type ClientProc func(*User, *Msg) int

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

	if err = websocket.JSON.Receive(conn, &p.Msg); err != nil {
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
func (c *Client) ProcMsg(msg *Msg) int {
	proc := procFuncArray[int(msg.Cmd)]
	if proc != nil {
		return proc(&c.User, msg)
	}

	// unknow cmd, kick client
	return PROC_KICK
}

func (c *Client) forwardRoutine() {
	var p *Packet
	var err error

	if c.scene == nil {
		fmt.Printf("no scene?\n")
		return
	}

	for {
		p, err = c.readPacket(c.conn)
		if err == nil {
			c.scene.cli_chan <- p
			continue
		}

		fmt.Printf("readpacket err, close client\n")

		// we have net or proto error
		// kick self
		c.KickPlayer(&c.User)
		break
	}
}

func (c *Client) detectDelay() error {
	var tv syscall.Timeval
	syscall.Gettimeofday(&tv)
	c.t1 = tv.Sec*1000000 + tv.Usec

	req := NewMsg()
	req.Cmd = kCmdDetectDelay
	c.Reply(req)

	p, err := c.readPacket(c.conn)
	if err != nil {
		return err
	}

	if p.Cmd != kCmdDetectDelay {
		fmt.Printf("detect time error, close client\n")
		return ErrDetectTime
	}

	if c.ProcMsg(&p.Msg) != PROC_OK {
		// login failed
		// should reply error
		fmt.Printf("detect time failed, close client\n")
		return ErrDetectTime
	}

	return nil
}

func (c *Client) Proc() {
	defer c.Close()

	var p *Packet

	// login
	p, err := c.readPacket(c.conn)
	if err != nil {
		return
	}

	if p.Cmd != kCmdUserLogin {
		fmt.Printf("not login, close client\n")
		return
	}

	if c.ProcMsg(&p.Msg) != PROC_OK {
		// login failed
		// should reply error
		fmt.Printf("login failed, close client\n")
		return
	}

	if c.detectDelay() != nil {
		return
	}

	fmt.Printf("%s login ok\n", c.userdb.Name)

	// proc player event message
	go c.UserEventRoutine()

	// wait msg and forward to scene
	c.forwardRoutine()

	fmt.Printf("%s logout\n", c.userdb.Name)
	return
}

func NewClient(conn *websocket.Conn) *Client {
	c := &Client{
		conn: conn,
	}
	InitUser(&c.User, c)
	return c
}

func InitClient(c *Client, conn *websocket.Conn) {
	c.conn = conn
	c.eventch = make(chan *Event, 128)
}
