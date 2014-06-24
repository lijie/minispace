package minispace

import "sync"
import "time"
import "fmt"

type Player interface {
	SendClient(msg *Msg) error
	SetUserId(id int)
	SetScene(s *Scene)
	UserId() int
	UserName() string
	SceneListNode() *List
	Status() *ShipStatus
	Update(delta float64)
	CheckHit(target Player) bool
	HpDown(value int) int
	Die()
	Beat()

	Position() (x int, y int)
}

// interal message
type Event struct {
	cmd int
	data interface{}
	sender Player
	callback func(*Event, error)
}

type Scene struct {
	activeList List
	deadList List
	num int
	cli_chan chan *Packet
	cmd_chan chan *Event
	lock sync.Mutex
	enable bool
	idmap int
	beamPool sync.Pool
}

var currentScene *Scene
func init() {
	currentScene = NewScene()
}

func CurrentScene() *Scene {
	return currentScene
}

func (s *Scene) freeId(id int) {
	if id > 16 {
		return
	}

	s.idmap = s.idmap &^ (1 << uint(id - 1))
}

func (s *Scene) allocId() (int, error) {
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

func (s *Scene) DelPlayer(u Player) {
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

	id := e.sender.UserId()
	// s.clientList.Remove(e.sender.pos)
	e.sender.SceneListNode().RemoveSelf()

	s.BroadProto(nil, true, kCmdUserKick, "id", &id)

	s.freeId(id)
	if e.callback != nil {
		e.callback(e, nil)
	}

	s.num--
}

func (s *Scene) AddPlayer(u Player) (chan *Packet, error) {
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

func (s *Scene) notifyProto(player Player, cmd float64, field string, data interface{}) {
	msg := NewMsg()
	msg.Cmd = cmd
	msg.Body[field] = data
	player.SendClient(msg)
}

func (s *Scene) BroadProto(sender Player, exclusion bool, cmd float64, field string, data interface{}) {
	var u Player

	msg := NewMsg()
	msg.Cmd = cmd
	msg.Body[field] = data

	for p := s.activeList.Next(); p != &s.activeList; p = p.Next() {
		u = p.Host().(Player)
		if u == sender && exclusion {
			continue
		}

		u.SendClient(msg)
	}
}

func (s *Scene) broadStopBeam(u Player, beamid int, hit int) {
	data := &ProtoStopBeam{
		Id: u.UserId(),
		BeamId: beamid,
		Hit: hit,
	}

	s.BroadProto(u, false, kCmdStopBeam, "data", data)
}

func (s *Scene) notifyAddUser(u Player) {
	var t Player
	var n []*ProtoAddUser
	for p := s.activeList.Next(); p != &s.activeList; p = p.Next() {
		t = p.Host().(Player)
		if t == u {
			continue
		}
		data := &ProtoAddUser{
			Id: t.UserId(),
			Name: t.UserName(),
		}
		n = append(n, data)
	}

	s.notifyProto(u, kCmdAddUser, "users", n)
}

func (s *Scene) broadAddUser(u Player) {
	var n []*ProtoAddUser
	data := &ProtoAddUser{
		Id: u.UserId(),
		Name: u.UserName(),
	}
	n = append(n, data)

	s.BroadProto(u, true, kCmdAddUser, "users", n)
}

func (s *Scene) addai(num int) {
	var ai *AIUser
	var id int
	var err error
	var l *List

	for i := 0; i < num; i++ {
		ai = NewAIUser()
		id, err = s.allocId()
		if err != nil {
			break
		}
		ai.SetScene(s)
		ai.SetUserId(id)
		l = ai.SceneListNode()
		s.activeList.PushBack(l)
		s.num++
		// tell all others I'm here
		s.broadAddUser(ai)
	}
}

func (s *Scene) addPlayer(e *Event) {
	e.sender.SetScene(s)

	if s.num >= 16 {
		e.data = false
		e.callback(e, nil)
		return
	}

	// each user have an unique id
	id, err := s.allocId()
	if err != nil {
		e.data = false
		e.callback(e, nil)
		return
	}

	fmt.Printf("alloc id %d for user %s\n", id, e.sender.UserName())
	// add ok
	e.sender.SetUserId(id)
	// e.sender.pos = s.clientList.PushBack(e.sender)

	// add to active list
	l := e.sender.SceneListNode()
	s.activeList.PushBack(l)

	s.num++
	e.data = true
	e.callback(e, nil)

	// tell all others I'm here
	s.broadAddUser(e.sender)

	// show enemies
	s.notifyAddUser(e.sender)

//	if s.num < 8 {
//		s.addai(8 - s.num)
//	}
}

func (s *Scene) broadShipDead(id int) {
	fmt.Printf("ship %d dead\n", id)
	s.BroadProto(nil, false, kCmdShipDead, "data", &id)
}

func (s *Scene) broadShipStatus() {
	if s.num == 0 {
		return
	}

	var c Player
	var n []*ShipStatus

	for p := s.activeList.Next(); p != &s.activeList; p = p.Next() {
		c = p.Host().(Player)
		n = append(n, c.Status())
	}

	s.BroadProto(nil, false, kCmdShipStatus, "users", n)
}

func (s *Scene) procTimeout() {
}

func checkHit(shooter Player, target Player) {
}

func (s *Scene) checkHitAll(shooter Player, l *List) {
	var tmp *List
	var hit bool
	var target Player
	var node *List

	p := l.Next()
	for p != l {
		tmp = p.Next()
		target = p.Host().(Player)
		p = tmp

		hit = shooter.CheckHit(target)
		if !hit {
			continue
		}

		// hit
		if target.HpDown(20) == 0 {
			s.broadShipDead(target.UserId())
			// remvoe target from active list
			node = target.SceneListNode()
			node.RemoveSelf()
			// add target to dead list
			s.deadList.PushBack(node)
			// notify target is dead
			target.Die()
		}
		// notify shooter
		shooter.Beat()
	}
}

func (s *Scene) runFrame(delta float64) {
	// update ship status for all clients
	s.broadShipStatus()

	// update for each user
	for p := s.activeList.Next(); p != &s.activeList; p = p.Next() {
		p.Host().(Player).Update(delta)
	}

	// check hit for each ship
	for p := s.activeList.Next(); p != &s.activeList; p = p.Next() {
		s.checkHitAll(p.Host().(Player), &s.activeList)
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
