// Copyright (c) LiJie 2014-05-23 17:32:46

package minispace

import "code.google.com/p/go.net/websocket"
import _ "time"
import "errors"
import "fmt"
import "sync"
import "container/list"

const (
	PROC_OK = 0
	PROC_ERR = 1
	PROC_KICK = 2
)

const (
	CCMD_KICK = 1
)

// Msg from client
type Msg struct {
	Cmd float64 `json:"cmd"`
	ErrCode float64 `json:"errcode"`
	Seq float64 `json:"seq"`
	Userid string `json:"userid"`
	Body map[string]interface{} `json:"body"`
}

// Msg wrapper for scene
type Packet struct {
	msg Msg
	client *Client
	ok bool
}

// Msg between client
type clientMsg struct {
	cmd int
	data interface{}
	sender *Client
	callback func(cmd *clientMsg)
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
	cmdch chan *clientMsg
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
		fmt.Printf("ClientMsg %#v\n", p.msg)
		p.ok = false
		return p, err
	} else {
		fmt.Printf("ClientMsg %#v\n", p.msg)
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
	proc := procFuncArray[int(msg.Cmd)]
	if proc != nil {
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

func (c *Client) KickName(name string) {
	other := SearchOnline(name)
	if other != nil {
		c.KickClient(other)
	}
}

func (c *Client) KickClient(other *Client) {
	cmd := &clientMsg{
		cmd: CCMD_KICK,
		sender: c,
	}

	var lock sync.Mutex
	lock.Lock()

	cmd.callback = func(cmd *clientMsg) {
		lock.Unlock()
	}

	fmt.Printf("here1\n")
	// send cmd
	other.cmdch <- cmd

	fmt.Printf("here2\n")
	// wait
	lock.Lock()
	fmt.Printf("here3\n")
}

func forwardRoutine(ch chan *Packet, c *Client) {
	var p *Packet
	var err error

	for {
		p, err = c.readPacket(c.conn)
		if err == nil && p.ok {
			ch <- p
			continue
		}

		// we have net or proto error
		// kick self
		c.KickClient(c)
		break
	}
}

func (c *Client) Proc() {
	defer c.Close()

	var p *Packet

	// login
	p, err := c.readPacket(c.conn)
	if err != nil || !p.ok {
		return
	}

	if p.msg.Cmd != kCmdUserLogin {
		fmt.Printf("not login, close client\n")
		return
	}

	if c.ProcMsg(&p.msg) != PROC_OK {
		// login failed
		// should reply error
		fmt.Printf("login failed, close client\n")
		return
	}

	fmt.Printf("login ok\n")
	// ok, login succ, add to a scene
	ch, err := CurrentScene().AddClient(c)
	if err != nil {
		fmt.Printf("add client err")
		return
	}

	fmt.Printf("add ok\n")
	c.insence = true

	// wait msg and forward to scene
	go forwardRoutine(ch, c)

	for c.enable {
		fmt.Printf("here4\n")
		// wait client cmd
		cmd := <- c.cmdch

		fmt.Printf("here5\n")
		// be kicked
		if cmd.cmd == CCMD_KICK {
			CurrentScene().DelClient(c)
			c.enable = false
			c.Logout()

			if cmd.callback != nil {
				fmt.Printf("here6\n")
				cmd.callback(cmd)
			}
		}
	}

	fmt.Printf("%s logout\n", c.Name)
	return
}

func NewClient(conn *websocket.Conn) *Client {
	c := &Client{
		conn: conn,
		enable: true,
		login: false,
		insence: false,
		cmdch: make(chan *clientMsg, 128),
	}
	InitUser(&c.User)
	return c
}
