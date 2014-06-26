package minispace

import _ "fmt"
import "math"

type AIAction interface {
	ActShoot() error
	ActMove(int)
	ActRotate(int)
}

type AIAlgo interface {
	Update(ai AIAction, dt float64)
	Name() string
}

type AIUser struct {
	Ship
	name string
	act int
	enable bool
	algo AIAlgo
	shipmap map[int]ShipStatus
	msgch chan *Msg
	eventch chan *Event
}

func (ai *AIUser) UserName() string {
	return ai.name
}

func (ai *AIUser) Die() {
	s := ai.scene

	// recover...
	ai.Hp = 100

	// remove self from dead list
	ai.sceneList.RemoveSelf()

	// push back to active list
	s.activeList.PushBack(&ai.sceneList)
}

func (ai *AIUser) Beat() {
}

func (ai *AIUser) updatePosition(delta float64) {
	if ai.Rotate == 2 {
		angle := ai.Angle + 80 * (delta / 1000);
		if angle >= 360 {
			angle = angle - 360;
		}
		ai.Angle = angle
	}

	if ai.Move == 1 {
		angle := ai.Angle + 90
		// move
		r := 80 * (delta / 1000);
		x := r * math.Sin(angle * math.Pi / 180);
		y := r * math.Cos(angle * math.Pi / 180);

		x = ai.X + x
		y = ai.Y + y

		if x > kScreenWidth {
			x = kScreenWidth
		} else if x < 0 {
			x = 0
		}

		if y > kScreenHeight {
			y = kScreenHeight
		} else if y < 0 {
			y = 0
		}

		ai.X = x
		ai.Y = y
	}
}

func (ai *AIUser) updateAction(delta float64) {
	if ai.act == 1 {
		// shoot
	}

	// clear
	ai.act = 0
}

func (ai *AIUser) Update(delta float64) {
	ai.algo.Update(ai, delta)

	ai.updateAction(delta)
	ShipUpdateBeam(ai, delta)
	ai.updatePosition(delta)
}

func (ai *AIUser) GetShip() *Ship {
	return &ai.Ship
}

func (ai *AIUser) SendClient(msg *Msg) error {
	ai.msgch <- msg
	return nil
}

// for AIAction
func (ai *AIUser) ActRotate(dir int) {
	ai.Rotate = dir
}

func (ai *AIUser) ActMove(dir int) {
	ai.Move = dir
}

func (ai *AIUser) ActShoot() error {
	mask := ai.beamMap
	id := 0

	for ((mask & 0x01) > 0) && id < 5 {
		id++
		mask = mask >> 1
	}

	if id >= 5 {
		return nil
	}

	ShipAddBeam(ai, id)
	return nil
}

func (ai *AIUser) procShipStatus(msg *Msg) {
	users, ok := msg.Body["users"].([]*ShipStatus)
	if !ok {
		return
	}

	if len(users) == 0 {
		return
	}

	for i := range users {
		// copy ship status
		ai.shipmap[users[i].Id] = *users[i]
	}
}

func (ai *AIUser) procMsg(msg *Msg) {
	if msg.Cmd == kCmdShipStatus {
		ai.procShipStatus(msg)
	}
}

func (ai *AIUser) procEvent(event *Event) {
	
}

func (ai *AIUser) aiEventRoutine() {
	var msg *Msg
	var event *Event

	for ai.enable {
		select {
		case msg =<- ai.msgch:
			ai.procMsg(msg)
		case event =<- ai.eventch:
			ai.procEvent(event)
		}
	}
		
}

func NewAIUser() *AIUser {
	ai := &AIUser{
		name: "AI",
		enable: true,
		shipmap: make(map[int]ShipStatus),
		msgch: make(chan *Msg, 128),
		eventch: make(chan *Event, 8),
	}
	InitShip(&ai.Ship)
	ai.X = 480
	ai.Y = 320
	ai.Hp = 100
	InitList(&ai.sceneList, ai)
	ai.algo = NewAISimapleAlgo()
	go ai.aiEventRoutine()
	return ai
}
