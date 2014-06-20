package minispace

import "sync"
import "time"
import "fmt"

type Scene struct {
	activeList List
	deadList List
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
	case kEventAddPlayer:
		s.addPlayer(e)
	case kEventDelPlayer:
		s.delPlayer(e)
	}
}

func (s *Scene) DelPlayer(u *User) {
	var lock sync.Mutex

	lock.Lock()
	cmd := Event{
		cmd: kEventDelPlayer,
		sender: u,
		callback: func(e *Event, err error) {
			lock.Unlock()
		},
	}

	s.cmd_chan <- &cmd
	// wait
	lock.Lock()
}

func (s *Scene) delPlayer(e *Event) {
	if s.num == 0 {
		e.callback(e, nil)
		return
	}

	id := e.sender.Id
	// s.clientList.Remove(e.sender.pos)
	e.sender.sceneList.RemoveSelf()

	s.BroadProto(nil, true, kCmdUserKick, "id", &id)

	s.setId(id)
	if e.callback != nil {
		e.callback(e, nil)
	}

	s.num--
}

func (s *Scene) AddPlayer(u *User) (chan *Packet, error) {
	var lock sync.Mutex
	var err error

	lock.Lock()
	cmd := Event{
		cmd: kEventAddPlayer,
		sender: u,
		callback: func(e *Event, err error) {
			succ := e.data.(bool)
			if !succ {
				err = ErrSceneFull
			}
			lock.Unlock()
		},
	}

	s.cmd_chan <- &cmd
//	fmt.Printf("wait add client\n")
	// wait
	lock.Lock()

	if err != nil {
		return nil, err
	}
	return s.cli_chan, nil
}

func (s *Scene) notifyProto(target *User, cmd float64, field string, data interface{}) {
	msg := NewMsg()
	msg.Cmd = cmd
	msg.Body[field] = data
	target.conn.Reply(msg)
}

func (s *Scene) BroadProto(sender *User, exclusion bool, cmd float64, field string, data interface{}) {
	var u *User

	msg := NewMsg()
	msg.Cmd = cmd
	msg.Body[field] = data

	for p := s.activeList.Next(); p != &s.activeList; p = p.Next() {
		u = p.Host().(*User)
		if !u.login {
			continue
		}
		if u == sender && exclusion {
			continue
		}

		u.conn.Reply(msg)
	}
}

func (s *Scene) broadStopBeam(u *User, beamid int, hit int) {
	data := &ProtoStopBeam{
		Id: u.Id,
		BeamId: beamid,
		Hit: hit,
	}

	s.BroadProto(u, false, kCmdStopBeam, "data", data)
}

func (s *Scene) notifyAddUser(u *User) {
	var t *User
	var n []*ProtoAddUser
	for p := s.activeList.Next(); p != &s.activeList; p = p.Next() {
		t = p.Host().(*User)
		if !t.login {
			continue
		}
		if t == u {
			continue
		}
		data := &ProtoAddUser{
			Id: t.Id,
			Name: t.Name,
		}
		n = append(n, data)
	}

	s.notifyProto(u, kCmdAddUser, "users", n)
}

func (s *Scene) broadAddUser(u *User) {
	var n []*ProtoAddUser
	data := &ProtoAddUser{
		Id: u.Id,
		Name: u.Name,
	}
	n = append(n, data)

	s.BroadProto(u, true, kCmdAddUser, "users", n)
}

func (s *Scene) addPlayer(e *Event) {
	e.sender.scene = s

	if s.num >= 16 {
		e.data = false
		e.callback(e, nil)
		return
	}

	// each user have an unique id
	id, err := s.getId()
	if err != nil {
		e.data = false
		e.callback(e, nil)
		return
	}

	fmt.Printf("alloc id %d for user %s\n", id, e.sender.Name)
	// add ok
	e.sender.Id = id
	// e.sender.pos = s.clientList.PushBack(e.sender)

	// add to active list
	s.activeList.PushBack(&e.sender.sceneList)

	s.num++
	e.data = true
	e.callback(e, nil)

	// tell all others I'm here
	s.broadAddUser(e.sender)

	// show enemies
	s.notifyAddUser(e.sender)
}

func (s *Scene) broadShipDead(id int) {
	fmt.Printf("ship %d dead\n", id)
	s.BroadProto(nil, false, kCmdShipDead, "data", &id)
}

func (s *Scene) broadShipStatus() {
	if s.num == 0 {
		return
	}

	var c *User
	var n []*ShipStatus

	for p := s.activeList.Next(); p != &s.activeList; p = p.Next() {
		c = p.Host().(*User)
		if !c.login {
			continue
		}
		if c.Hp == 0 {
			continue
		}
		n = append(n, &c.ShipStatus)
	}

	s.BroadProto(nil, false, kCmdShipStatus, "users", n)
}

func (s *Scene) procTimeout() {
}

func (s *Scene) runFrame(delta float64) {
	// update ship status for all clients
	s.broadShipStatus()

	// update for each user
	for p := s.activeList.Next(); p != &s.activeList; p = p.Next() {
		p.Host().(*User).Update(delta, s)
	}

	// check hit for each ship
	for p := s.activeList.Next(); p != &s.activeList; p = p.Next() {
		p.Host().(*User).CheckHitAll(&s.activeList, s)
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
	s := &Scene{
		enable: true,
		cli_chan: make(chan *Packet, 1024),
		cmd_chan: make(chan *Event, 128),
	}
	InitList(&s.activeList, s)
	InitList(&s.deadList, s)
	return s
}
