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
	enable bool
	algo AIAlgo
	shipmap map[int]ShipStatus
	msgch chan *Msg
	eventch chan *Event
}

func (ai *AIUser) UserName() string {
	return ai.name
}

func (ai *AIUser) Beat() {
}

func (ai *AIUser) doMove(dt float64) {
	angle := ai.ship.Angle + 90
	// move
	r := 80 * (dt / 1000);
	x := r * math.Sin(angle * math.Pi / 180);
	y := r * math.Cos(angle * math.Pi / 180);

	// forward
	if ai.ship.Move == 1 {
		x = ai.ship.X + x
		y = ai.ship.Y + y
	} else {
		// backward
		x = ai.ship.X - x
		y = ai.ship.Y - y
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

	ai.ship.X = x
	ai.ship.Y = y
}

func (ai *AIUser) doRotate(dt float64) {
	var angle float64

	if ai.ship.Rotate == 2 {
		angle = ai.ship.Angle + 120 * (dt / 1000);
	} else {
		angle = ai.ship.Angle - 120 * (dt / 1000);
	}
	if angle >= 360 {
		angle = angle - 360
	}
	if angle < 0 {
		angle = angle + 360
	}

	ai.ship.Angle = angle
}

func (ai *AIUser) updatePosition(delta float64) {
	ai.doRotate(delta)
	ai.doMove(delta)
}

func (ai *AIUser) Update(delta float64) {
	if ai.status == kStatusActive {
		ai.algo.Update(ai, delta)
		ShipUpdateBeam(ai, delta)
		ai.updatePosition(delta)
	}
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
	ai.ship.Rotate = dir
}

func (ai *AIUser) ActMove(dir int) {
	ai.ship.Move = dir
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
	ai.ship.X = 480
	ai.ship.Y = 320
	ai.ship.Hp = 100
	InitList(&ai.sceneList, ai)
	InitList(&ai.statusList, ai)
	ai.algo = NewAISimapleAlgo()
	go ai.aiEventRoutine()
	return ai
}
