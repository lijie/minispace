package minispace

import "sync"
import "errors"
import "time"
import _ "fmt"

type Scene struct {
	clients [16]*Client
	num int
	cli_chan chan *Packet
	lock sync.Mutex
	enable bool
}

type protoUser struct {
	X int `json:"x"`
	Y int `json:"y"`
	RO int `json:"ro"`
	ID int `json:"id"`
}

type protoUserNotify struct {
	Users []protoUser
}

var currentScene *Scene
func init() {
	currentScene = NewScene()
}

func CurrentScene() *Scene {
	return currentScene
}

func (s *Scene) DelClient(c *Client) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.num == 0 {
		return
	}

	i := 0
	for ; i < 16; i++ {
		if s.clients[i] == c {
			s.clients[i] = nil
			s.num--
			break
		}
	}

	if i >= 16 {
		return
	}

	reply := NewMsg()
	reply.Cmd = kCmdUserKick
	reply.Body["id"] = i
	s.notifyAll(reply)
	return
}

func (s *Scene) AddClient(c *Client) (chan *Packet, error) {
	c.scene = s

	s.lock.Lock()
	defer s.lock.Unlock()

	if s.num >= 16 {
		return nil, errors.New("Scene is full")
	}

	i := 0
	for ; i < 16; i++ {
		if s.clients[i] == nil {
			s.clients[i] = c
			c.id = i
			break
		}
	}

	if i >= 16 {
		return nil, errors.New("Scene is full")
	}

	s.num++
	return s.cli_chan, nil
}

func (s *Scene) notifyAll(msg *Msg) {
	for i := 0; i < 16; i++ {
		if s.clients[i] == nil {
			continue
		}
		s.clients[i].Reply(msg)
	}
}

func (s *Scene) updateAll() {
	if s.num == 0 {
		return
	}

	var p protoUser
	var c *Client
	var n []protoUser
	
	for i := 0; i < 16; i++ {
		if s.clients[i] == nil {
			continue
		}

		c = s.clients[i]
		if !c.login {
			continue
		}

		p.X = c.x
		p.Y = c.y
		p.RO = c.ro
		p.ID = c.id

		n = append(n, p)
	}

	msg := NewMsg()
	msg.Cmd = kCmdUserNotify
	msg.Body["users"] = n

	s.notifyAll(msg)
}

func (s *Scene) procTimeout() {
}

func (s *Scene) Run() {
	timer_ch := make(chan int, 1)

	go func() {
		for s.enable {
			<-time.After(30 * time.Millisecond)
			timer_ch <- 1
		}
	}()

	var p *Packet
	for s.enable {
		select {
		case p = <-s.cli_chan:
			if p != nil {
				if p.ok {
					p.client.ProcMsg(&p.msg)
				} else {
					s.DelClient(p.client)
				}
			}
		case _ = <- timer_ch:
			s.updateAll()
		}
	}
}

func NewScene() *Scene {
	return &Scene{
		enable: true,
		cli_chan: make(chan *Packet, 1024),
	}
}
