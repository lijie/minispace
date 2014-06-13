package minispace

import "sync"
import "time"
import "fmt"
import "container/list"

const (
	SCENE_ADD_CLIENT = 1
	SCENE_DEL_CLIENT = 2
)

type sceneCmd struct {
	cmd int
	c *Client
	callback func(cmd int, c *Client, succ bool)
}

type Scene struct {
	clientList *list.List
	num int
	cli_chan chan *Packet
	cmd_chan chan *sceneCmd
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
	if id >= 16 {
		return
	}

	s.idmap = s.idmap &^ (1 << uint(id))
}

func (s *Scene) getId() (int, error) {
	if s.num >= 16 {
		return -1, ErrSceneFull
	}

	tmp := s.idmap
	for i := 0; i < 16; i++ {
		if tmp & 0x01 == 0 {
			s.idmap = s.idmap | (1 << uint(i))
			return i, nil
		}
		tmp = tmp >> 1
	}

	return -1, ErrSceneFull
}

func (s *Scene) doSceneCmd(cmd *sceneCmd) {
	switch cmd.cmd {
	case SCENE_ADD_CLIENT:
		s.addClient(cmd)
	case SCENE_DEL_CLIENT:
		s.delClient(cmd)
	}
}

func (s *Scene) DelClient(c *Client) {
	var lock sync.Mutex

	lock.Lock()
	cmd := sceneCmd{
		cmd: SCENE_DEL_CLIENT,
		c: c,
		callback: func(cmd int, c *Client, succ bool) {
			lock.Unlock()
		},
	}

	s.cmd_chan <- &cmd
	// wait
	lock.Lock()
}

func (s *Scene) delClient(cmd *sceneCmd) {
	if s.num == 0 {
		cmd.callback(SCENE_DEL_CLIENT, cmd.c, true)
		return
	}

	id := cmd.c.Id
	s.clientList.Remove(cmd.c.pos)

	reply := NewMsg()
	reply.Cmd = kCmdUserKick
	reply.Body["id"] = id
	s.notifyAll(reply)

	s.setId(id)
	cmd.callback(SCENE_DEL_CLIENT, cmd.c, true)
}

func (s *Scene) AddClient(c *Client) (chan *Packet, error) {
	var lock sync.Mutex
	var err error

	lock.Lock()
	cmd := sceneCmd{
		cmd: SCENE_ADD_CLIENT,
		c: c,
		callback: func(cmd int, c *Client, succ bool) {
			if !succ {
				err = ErrSceneFull
			}
			lock.Unlock()
		},
	}

	s.cmd_chan <- &cmd
	fmt.Printf("wati add client\n")
	// wait
	lock.Lock()

	if err != nil {
		return nil, err
	}
	return s.cli_chan, nil
}

func (s *Scene) addClient(cmd *sceneCmd) {
	cmd.c.scene = s

	if s.num >= 16 {
		cmd.callback(SCENE_ADD_CLIENT, cmd.c, false)
		return
	}

	id, err := s.getId()
	if err != nil {
		cmd.callback(SCENE_ADD_CLIENT, cmd.c, false)
		return
	}

	cmd.c.Id = id
	cmd.c.pos = s.clientList.PushBack(cmd.c)
	s.num++
	cmd.callback(SCENE_ADD_CLIENT, cmd.c, true)
}

func (s *Scene) notifyAll(msg *Msg) {
	for p := s.clientList.Front(); p != nil; p = p.Next() {
		p.Value.(*Client).Reply(msg)
	}
}

func (s *Scene) updateAll() {
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
		// clear action
		c.Act = 0
	}

	msg := NewMsg()
	msg.Cmd = kCmdUserNotify
	msg.Body["users"] = n

	s.notifyAll(msg)
}

func (s *Scene) procTimeout() {
}

func (s *Scene) runFrame(delta float64) {
	// update status for all clients
	s.updateAll()

	// update for each user
	for p := s.clientList.Front(); p != nil; p = p.Next() {
		p.Value.(*Client).Update(delta)
	}

	// check hit for each ship
	for p := s.clientList.Front(); p != nil; p = p.Next() {
		p.Value.(*Client).CheckHitAll(s.clientList)
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
	var cmd *sceneCmd

	for s.enable {
		select {
		case p = <-s.cli_chan:
			p.client.ProcMsg(&p.msg)
		case _ = <- timer_ch:
			s.runFrame(50.0)
		case cmd = <- s.cmd_chan:
			s.doSceneCmd(cmd)
		}
	}
}

func NewScene() *Scene {
	return &Scene{
		enable: true,
		cli_chan: make(chan *Packet, 1024),
		cmd_chan: make(chan *sceneCmd, 128),
		clientList: list.New(),
	}
}
