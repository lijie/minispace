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
	name    string
	enable  bool
	algo    AIAlgo
	shipmap map[int]ShipStatus
	msgch   chan *Msg
	eventch chan *Event
}

func (ai *AIUser) doMove(dt float64) {
	angle := ai.status.Angle + 90
	// move
	r := kShipSpeed * (dt / 1000)
	x := r * math.Sin(angle*math.Pi/180)
	y := r * math.Cos(angle*math.Pi/180)

	// forward
	if ai.status.Move == 1 {
		x = ai.status.X + x
		y = ai.status.Y + y
	} else {
		// backward
		x = ai.status.X - x
		y = ai.status.Y - y
	}

	if x > kMapWidth {
		x = kMapWidth
	} else if x < 0 {
		x = 0
	}

	if y > kMapHeight {
		y = kMapHeight
	} else if y < 0 {
		y = 0
	}

	ai.status.X = x
	ai.status.Y = y
}

func (ai *AIUser) doRotate(dt float64) {
	var angle float64

	if ai.status.Rotate == 2 {
		angle = ai.status.Angle + 120*(dt/1000)
	} else {
		angle = ai.status.Angle - 120*(dt/1000)
	}
	if angle >= 360 {
		angle = angle - 360
	}
	if angle < 0 {
		angle = angle + 360
	}

	ai.status.Angle = angle
}

func (ai *AIUser) updatePosition(delta float64) {
	ai.doRotate(delta)
	ai.doMove(delta)
}

// for Player
func (ai *AIUser) Update(delta float64) {
	if ai.state == kStateActive {
		ai.algo.Update(ai, delta)
		ai.updatePosition(delta)
	}
}

func (ai *AIUser) SendMsg(msg *Msg) error {
	ai.msgch <- msg
	return nil
}

func (ai *AIUser) Name() string {
	return ai.name
}

func (ai *AIUser) Win() {
}

func (ai *AIUser) Dead() {
}

// for AIAction
func (ai *AIUser) ActRotate(dir int) {
	ai.status.Rotate = dir
}

func (ai *AIUser) ActMove(dir int) {
	ai.status.Move = dir
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

	ai.AddBeam(id)
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
		case msg = <-ai.msgch:
			ai.procMsg(msg)
		case event = <-ai.eventch:
			ai.procEvent(event)
		}
	}

}

func NewAIUser() *AIUser {
	ai := &AIUser{
		name:    "AI",
		enable:  true,
		shipmap: make(map[int]ShipStatus),
		msgch:   make(chan *Msg, 128),
		eventch: make(chan *Event, 8),
	}
	InitShip(&ai.Ship, ai)
	ai.status.X = 480
	ai.status.Y = 320
	ai.status.Hp = 100
	ai.algo = NewAISimapleAlgo()
	go ai.aiEventRoutine()
	return ai
}
