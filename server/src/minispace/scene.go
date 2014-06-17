package minispace

import "sync"
import "time"
import "fmt"
import "container/list"

type Scene struct {
	clientList *list.List
	num int
	cli_chan chan *Packet
	cmd_chan chan *Event
	lock sync.Mutex
	enable bool
	idmap int
}

var currentScene *Scene
func init() {
	currentScene = NewScene()
}

func CurrentScene() *Scene {
	return currentScene
}

func (s *Scene) setId(id int) {
	if id > 16 {
		return
	}

	s.idmap = s.idmap &^ (1 << uint(id - 1))
}

func (s *Scene) getId() (int, error) {
	if s.num >= 16 {
		return -1, ErrSceneFull
	}

	tmp := s.idmap
	for i := 0; i < 16; i++ {
		if tmp & 0x01 == 0 {
			s.idmap = s.idmap | (1 << uint(i))
			return (i + 1), nil
		}
		tmp = tmp >> 1
	}

	return -1, ErrSceneFull
}

func (s *Scene) doSceneEvent(e *Event) {
	switch e.cmd {
	case kEventAddClient:
		s.addClient(e)
	case kEventDelClient:
		s.delClient(e)
	}
}

func (s *Scene) DelClient(c *Client) {
	var lock sync.Mutex

	lock.Lock()
	cmd := Event{
		cmd: kEventDelClient,
		sender: c,
		callback: func(e *Event) {
			lock.Unlock()
		},
	}

	s.cmd_chan <- &cmd
	// wait
	lock.Lock()
}

func (s *Scene) delClient(e *Event) {
	if s.num == 0 {
		e.callback(e)
		return
	}

	id := e.sender.Id
	s.clientList.Remove(e.sender.pos)

	s.BroadProto(nil, true, kCmdUserKick, "id", &id)

	s.setId(id)
	if e.callback != nil {
		e.callback(e)
	}

	s.num--
}

func (s *Scene) AddClient(c *Client) (chan *Packet, error) {
	var lock sync.Mutex
	var err error

	lock.Lock()
	cmd := Event{
		cmd: kEventAddClient,
		sender: c,
		callback: func(e *Event) {
			succ := e.data.(bool)
			if !succ {
				err = ErrSceneFull
			}
			lock.Unlock()
		},
	}

	s.cmd_chan <- &cmd
//	fmt.Printf("wati add client\n")
	// wait
	lock.Lock()

	if err != nil {
		return nil, err
	}
	return s.cli_chan, nil
}

type ProtoAddUser struct {
	Id int `json:"id"`
	Name string `json:"name"`
}

type ProtoStopBeam struct {
	Id int `json:"id"`
	BeamId int `json:"beamid"`
	Hit int `json:"hit"`
}

func (s *Scene) notifyProto(target *Client, cmd float64, field string, data interface{}) {
	msg := NewMsg()
	msg.Cmd = cmd
	msg.Body[field] = data
	target.Reply(msg)
}

func (s *Scene) BroadProto(sender *Client, exclusion bool, cmd float64, field string, data interface{}) {
	var c *Client

	msg := NewMsg()
	msg.Cmd = cmd
	msg.Body[field] = data

	for p := s.clientList.Front(); p != nil; p = p.Next() {
		c = p.Value.(*Client)
		if !c.login {
			continue
		}
		if c == sender && exclusion {
			continue
		}

		c.Reply(msg)
	}
}

func (s *Scene) broadStopBeam(c *Client, beamid int, hit int) {
	data := &ProtoStopBeam{
		Id: c.Id,
		BeamId: beamid,
		Hit: hit,
	}

	s.BroadProto(c, false, kCmdStopBeam, "data", data)
}

func (s *Scene) notifyAddUser(c *Client) {
	var t *Client
	var n []*ProtoAddUser
	for p := s.clientList.Front(); p != nil; p = p.Next() {
		t = p.Value.(*Client)
		if !t.login {
			continue
		}
		if t == c {
			continue
		}
		data := &ProtoAddUser{
			Id: t.Id,
			Name: t.Name,
		}
		n = append(n, data)
	}

	s.notifyProto(c, kCmdAddUser, "users", n)
}

func (s *Scene) broadAddUser(c *Client) {
	var n []*ProtoAddUser
	data := &ProtoAddUser{
		Id: c.Id,
		Name: c.Name,
	}
	n = append(n, data)

	s.BroadProto(c, true, kCmdAddUser, "users", n)
}

func (s *Scene) addClient(e *Event) {
	e.sender.scene = s

	if s.num >= 16 {
		e.data = false
		e.callback(e)
		return
	}

	// each user have an unique id
	id, err := s.getId()
	if err != nil {
		e.data = false
		e.callback(e)
		return
	}

	fmt.Printf("alloc id %d for user %s\n", id, e.sender.Name)
	// add ok
	e.sender.Id = id
	e.sender.pos = s.clientList.PushBack(e.sender)
	s.num++
	e.data = true
	e.callback(e)

	// tell all others I'm here
	s.broadAddUser(e.sender)

	// show enemies
	s.notifyAddUser(e.sender)
}

func (s *Scene) broadShipStatus() {
	if s.num == 0 {
		return
	}

	var c *Client
	var n []*ShipStatus

	for p := s.clientList.Front(); p != nil; p = p.Next() {
		c = p.Value.(*Client)
		if !c.login {
			continue
		}
		n = append(n, &c.User.ShipStatus)
	}

	s.BroadProto(nil, false, kCmdShipStatus, "users", n)
}

func (s *Scene) procTimeout() {
}

func (s *Scene) runFrame(delta float64) {
	// update ship status for all clients
	s.broadShipStatus()

	// update for each user
	for p := s.clientList.Front(); p != nil; p = p.Next() {
		p.Value.(*Client).Update(delta)
	}

	// check hit for each ship
	for p := s.clientList.Front(); p != nil; p = p.Next() {
		p.Value.(*Client).CheckHitAll(s.clientList, s)
	}
}

func (s *Scene) Run() {
	timer_ch := make(chan int, 1)

	go func() {
		for s.enable {
			<-time.After(50 * time.Millisecond)
			timer_ch <- 1
		}
	}()

	var p *Packet
	var e *Event

	for s.enable {
		select {
		case p = <-s.cli_chan:
			p.client.ProcMsg(&p.Msg)
		case _ = <- timer_ch:
			s.runFrame(50.0)
		case e = <- s.cmd_chan:
			s.doSceneEvent(e)
		}
	}
}

func NewScene() *Scene {
	return &Scene{
		enable: true,
		cli_chan: make(chan *Packet, 1024),
		cmd_chan: make(chan *Event, 128),
		clientList: list.New(),
	}
}
