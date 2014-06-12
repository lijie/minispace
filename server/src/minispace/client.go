// Copyright (c) LiJie 2014-05-23 17:32:46

package minispace

import "code.google.com/p/go.net/websocket"
import _ "time"
import "errors"
import "fmt"
import "container/list"

const (
	PROC_OK = 0
	PROC_ERR = 1
	PROC_KICK = 2
)

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
	insence bool
	scene *Scene
	lasterr int
	pos *list.Element
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

func (c *Client) SetErrCode(code int) {
	c.lasterr = code
}

func (c *Client) Reply(msg *Msg) {
//	fmt.Printf("reply %v\n", msg)
	websocket.JSON.Send(c.conn, msg)
}

// called in scene goroutine
func (c *Client) ProcMsg(msg *Msg) int {
//	fmt.Println(*msg)
	proc := procFuncArray[int(msg.Cmd)]
	if proc != nil {
//		fmt.Printf("proc %v\n", msg)
		return proc(c, msg)
	}

	// unknow cmd, kick client
	return PROC_KICK
}

func (c *Client) procTimeout() {
	c.enable = false
}

func (c *Client) Kick() {
	c.enable = false
}

func (c *Client) Proc() {
	defer c.Close()

	var p *Packet

	// login
	p, err = c.readPacket(c.conn)
	if err != nil || !p.ok {
		return
	}

	if p.msg.Cmd != kCmdUserLogin {
		return
	}

	if c.ProcMsg(&p.msg) != PROC_OK {
		// login failed
		// should reply error
		return
	}

	// ok, login succ, add to a scene
	ch, err = CurrentScene().AddClient(c)
	if err != nil {
		fmt.Printf("add client err")
		return
	}

	c.insence = true

	// forward msg to scene routine
	for c.enable {
		p, err = c.readPacket(c.conn)
		if err == nil && p.ok {
			ch <- p
		} else {
			CurrentScene().DelClient(c)
			c.enable = false
		}
	}

	return
}

func NewClient(conn *websocket.Conn) *Client {
	c := &Client{
		conn: conn,
		enable: true,
		login: false,
		insence: false,
	}
	InitUser(&c.User)
	return c
}
