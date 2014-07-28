package minispace

import "sync"
import "time"
import "fmt"

// interal message
type Event struct {
	cmd int
	data interface{}
	sender *Ship
	callback func(*Event, error)
}

type Scene struct {
	// save all ships in this scene
	shipList List
	// all alive ships
	activeList List
	// all dead ships
	deadList List
	// ship count
	num int
	// bitmap for ship id
	idmap int
	enable bool
	waitStart time.Time
	// packet from ship
	cli_chan chan *Packet
	// for internal event
	cmd_chan chan *Event
}

// TODO:
// only one scene right now
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
	case kEventRunFrame:
		dt := time.Now().Sub(s.waitStart).Nanoseconds()
		fmt.Printf("run frame by event, dt %d\n", dt)
		// s.runFrame(float64(dt))
	}
}

func (s *Scene) DelPlayer(u *Ship) {
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
	ship := e.sender
	ship.sceneList.RemoveSelf()
	ship.stateList.RemoveSelf()

	s.BroadProto(nil, true, kCmdUserKick, "id", &id)

	s.freeId(id)
	if e.callback != nil {
		e.callback(e, nil)
	}

	s.num--
}

func (s *Scene) StartRunFrame() {
	cmd := &Event{
		cmd: kEventRunFrame,
	}
	s.cmd_chan <- cmd
}

func (s *Scene) AddPlayer(u *Ship) (chan *Packet, error) {
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
	fmt.Printf("wait add client\n")
	// wait
	lock.Lock()
	fmt.Printf("wait add client done\n")

	if err != nil {
		return nil, err
	}
	return s.cli_chan, nil
}

func (s *Scene) notifyProto(ship *Ship, cmd float64, field string, data interface{}) {
	msg := NewMsg()
	msg.Cmd = cmd
	msg.Body[field] = data
	ship.SendMsg(msg)
}

func (s *Scene) BroadProto(sender *Ship, exclusion bool, cmd float64, field string, data interface{}) {
	var u *Ship

	msg := NewMsg()
	msg.Cmd = cmd
	msg.Body[field] = data

	for p := s.shipList.Next(); p != &s.shipList; p = p.Next() {
		u = p.Host().(*Ship)
		if u == sender && exclusion {
			continue
		}

		u.SendMsg(msg)
	}
}

func (s *Scene) broadStopBeam(u *Ship, beamid int, hit int) {
	data := &ProtoStopBeam{
		Id: u.UserId(),
		BeamId: beamid,
		Hit: hit,
	}

	s.BroadProto(u, false, kCmdStopBeam, "data", data)
}

func (s *Scene) notifyAddUser(u *Ship) {
	var t *Ship
	var n []*ProtoAddUser
	for p := s.shipList.Next(); p != &s.shipList; p = p.Next() {
		t = p.Host().(*Ship)
		if t == u {
			continue
		}
		data := &ProtoAddUser{
			Id: t.UserId(),
			Name: t.Name(),
		}
		n = append(n, data)
	}

	s.notifyProto(u, kCmdAddUser, "users", n)
}

func (s *Scene) broadAddUser(u *Ship) {
	var n []*ProtoAddUser
	data := &ProtoAddUser{
		Id: u.UserId(),
		Name: u.Name(),
	}
	n = append(n, data)

	s.BroadProto(u, true, kCmdAddUser, "users", n)
}

func (s *Scene) addai(num int) {
	var ai *AIUser
	var id int
	var err error
	var ship *Ship

	for i := 0; i < num; i++ {
		ai = NewAIUser()
		id, err = s.allocId()
		if err != nil {
			break
		}

		ai.SetScene(s)
		ai.SetUserId(id)

		ship = &ai.Ship

		s.activeList.PushBack(&ship.stateList)
		s.shipList.PushBack(&ship.sceneList)

		s.num++
		fmt.Printf("add ai %d\n", id)

		// tell all others I'm here
		s.broadAddUser(ship)
	}
}

func (s *Scene) addPlayer(e *Event) {
	fmt.Printf("call addPlayer\n")
	e.sender.SetScene(s)

	if s.num >= 16 {
		e.data = false
		e.callback(e, nil)
		return
	}

	fmt.Printf("call addPlayer 1\n")
	// each user have an unique id
	id, err := s.allocId()
	if err != nil {
		e.data = false
		e.callback(e, nil)
		return
	}

	fmt.Printf("alloc id %d for user %s\n", id, e.sender.Name())
	// add ok
	e.sender.SetUserId(id)
	// e.sender.pos = s.clientList.PushBack(e.sender)

	// add to active list
	ship := e.sender
	s.activeList.PushBack(&ship.stateList)
	s.shipList.PushBack(&ship.sceneList)

	s.num++
	e.data = true
	e.callback(e, nil)

	// tell all others I'm here
	s.broadAddUser(e.sender)

	// show enemies
	s.notifyAddUser(e.sender)

	if miniConfig.EnableAI && s.num < 8 {
		fmt.Printf("add %d ai\n", 8 - s.num)
		s.addai(8 - s.num)
		// s.addai(1)
	}
}

func (s *Scene) broadShipDead(id int) {
	fmt.Printf("ship %d dead\n", id)
	s.BroadProto(nil, false, kCmdShipDead, "data", &id)
}

func (s *Scene) broadShipStatus() {
	if s.num == 0 {
		return
	}

	var c *Ship
	var n []*ShipStatus

	for p := s.shipList.Next(); p != &s.shipList; p = p.Next() {
		c = p.Host().(*Ship)
		if c.updated {
			n = append(n, c.Status())
			c.updated = false
		}
	}

	if len(n) == 0 {
		return
	}

	s.BroadProto(nil, false, kCmdShipStatus, "users", n)
}

func (s *Scene) procTimeout() {
}

func (s *Scene) checkHitAll(shooter *Ship, l *List) {
	var tmp *List
	var hit bool
	var target *Ship

	p := l.Next()
	for p != l {
		tmp = p.Next()
		target = p.Host().(*Ship)
		p = tmp

		hit = ShipCheckHit(shooter, target)
		if !hit {
			continue
		}

		// hit
		if target.HpDown(20) == 0 {
			// remvoe target from active list
			target.stateList.RemoveSelf()

			// add target to dead list
			s.deadList.PushBack(&target.stateList)
			// notify target is dead
			target.Dead()
			s.broadShipDead(target.UserId())
		}
		// notify shooter
		shooter.Win()
	}
}

func (s *Scene) checkDead(ship *Ship, delta float64) {
	ship.deadCD += delta

//	if ship.deadCD > 5000 {
//		// restart
//		ship.deadCD = 0
//		ship.status.Hp = 100
//		ship.stateList.RemoveSelf()
//		s.activeList.PushBack(&ship.stateList)
//		ship.SetActive()
//		ship.scene.BroadProto(ship, false, kCmdShipRestart, "data", ship.status.Id)
//		fmt.Printf("ship %d restart\n", ship.status.Id)
//	}
}

func (s *Scene) runFrame(delta float64) {
	// update ship status for all clients
	s.broadShipStatus()

	// update for each user
	for p := s.activeList.Next(); p != &s.activeList; p = p.Next() {
		p.Host().(*Ship).Update(delta)
	}

	// check hit for each ship
	for p := s.activeList.Next(); p != &s.activeList; p = p.Next() {
		s.checkHitAll(p.Host().(*Ship), &s.activeList)
	}

	// update dead list
	var tmp *List
	var target *Ship
	p := s.deadList.Next()
	for p != &s.deadList {
		tmp = p.Next()
		target = p.Host().(*Ship)
		p = tmp

		s.checkDead(target, delta)
	}
}

const (
	SceneSleepTime = 50
)

func (s *Scene) Run() {
	timer_ch := make(chan int, 1)

	go func() {
		for s.enable {
			<-time.After(SceneSleepTime * time.Millisecond)
			timer_ch <- 1
		}
	}()

	var p *Packet
	var e *Event

	tickch := time.Tick(SceneSleepTime * time.Millisecond)

	for s.enable {
		s.waitStart = time.Now()
		select {
		case p = <-s.cli_chan:
			p.client.ProcMsg(&p.Msg)
		case _ = <- tickch:
			s.runFrame(float64(SceneSleepTime))
		case e = <- s.cmd_chan:
			s.doSceneEvent(e)
		}
	}
	fmt.Printf("scene stop\n")
}

func NewScene() *Scene {
	s := &Scene{
		enable: true,
		cli_chan: make(chan *Packet, 1024),
		cmd_chan: make(chan *Event, 128),
	}
	InitList(&s.activeList, s)
	InitList(&s.deadList, s)
	InitList(&s.shipList, s)
	return s
}
