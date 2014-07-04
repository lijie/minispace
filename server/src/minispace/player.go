package minispace

import "fmt"
import "math"
import "container/list"

// better Player interface
type Player2 interface {
	SendMsg(msg *Msg) error
	Update(dt float64)
	Name() string
	Dead()
	Win()
}

type ShipStatus struct {
	// position
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Angle float64 `json:"angle"`
	// forward or backward
	Move int `json:"move"`
	// left-rotate or right->rotate
	Rotate int `json:"rotate"`
	// shoot
	// Act int `json:"act"`
	Hp float64 `json:"hp"`
	// ship id
	Id int `json:"id"`
}

type Ship struct {
	Player2
	status ShipStatus
	// beam manager
	beamMap int
	beamList *list.List
	// which scene we belong to
	scene *Scene
	sceneList List
	stateList List
	deadCD float64
	state int
}

type Rect struct {
	x, y, width, height int
}

func (r *Rect) InRect(x, y int) bool {
	if x < r.x || x > r.x + r.width {
		return false
	}
	if y < r.y || y > r.y + r.height {
		return false
	}
	return true
}

type Beam struct {
	X, Y, Angle float64
	radian float64
	distance float64
	id int
	pos *list.Element
}

func (b *Beam) Hit(x int, y int) bool {
	r := Rect{x - 25, y - 25, 50, 50}
	if r.InRect(int(b.X), int(b.Y)) {
		return true
	}
	return false
}

// beam speed: 1000pix/3seconds
func (b *Beam) Update(delta float64) bool {
	// update position
	r := 1000.0 / (3.0 * 1000.0) * delta;
	b.X = b.X + r * math.Sin(b.radian)
	b.Y = b.Y + r * math.Cos(b.radian)
//	fmt.Printf("beam XY: %f, %f, %f\n", b.X, b.Y, r)
	b.distance += r

	if b.distance > 1000 {
		return false
	}
	// if out of screen?
	if b.X < 0 || b.X > kMapWidth {
		return false
	}
	if b.Y < 0 || b.Y > kMapHeight {
		return false
	}
	return true
}

func (ship *Ship) SendMsg(msg *Msg) error {
	return ship.Player2.SendMsg(msg)
}

func (ship *Ship) Update(dt float64) {
	ship.updateBeam(dt)
	ship.Player2.Update(dt)
}

func (ship *Ship) Name() string {
	return ship.Player2.Name()
}

func (p *Ship) SetActive() {
	p.state = kStateActive
}

func (p *Ship) Dead() {
	p.state = kStateDead
	p.Player2.Dead()
}

func (p *Ship) Win() {
	p.Player2.Win()
}

func (p *Ship) SetScene(s *Scene) {
	p.scene = s
}

func (p *Ship) SetUserId(id int) {
	p.status.Id = id
}

func (p *Ship) UserId() int {
	return p.status.Id
}

func (p *Ship) HpDown(value int) int {
	p.status.Hp -= float64(value)
	if p.status.Hp < 0 {
		p.status.Hp = 0
	}
	return int(p.status.Hp)
}

func (p *Ship) Status() *ShipStatus {
	return &p.status
}

func ShipCheckHit(sender *Ship, target *Ship) bool {
	if sender.status.Id == target.status.Id {
		return false
	}

	var beam *Beam
	for b := sender.beamList.Front(); b != nil; b = b.Next() {
		beam = b.Value.(*Beam)
		// TODO use float64 defalut
		if !beam.Hit(int(target.status.X), int(target.status.Y)) {
			continue
		}

		fmt.Printf("%d hit target %d\n", sender.status.Id, target.status.Id)

		sender.beamList.Remove(b)
		sender.beamMap = sender.beamMap &^ (1 << uint(beam.id))
		sender.scene.broadStopBeam(sender, int(beam.id), 1)

		// BUG:
		return true
	}

	return false
}

func (ship *Ship) updateBeam(dt float64) {
	var tmp *list.Element
	var beam *Beam

	for b := ship.beamList.Front(); b != nil; {
		beam = b.Value.(*Beam)
		tmp = b.Next()

		if !beam.Update(dt) {
			ship.beamList.Remove(b)
			ship.beamMap = ship.beamMap &^ (1 << uint(beam.id))
			ship.scene.broadStopBeam(ship, int(beam.id), 0)
		}

		b = tmp
	}
}

func (ship *Ship) AddBeam(beamid int) error {
	// check beamid is valid
	if ((1 << uint(beamid)) & ship.beamMap) != 0 {
		// error
		fmt.Printf("beamid %d already used\n", beamid)
		return ErrInvalidBeamID
	}

	// save beamid and beam
	ship.beamMap |= (1 << uint(beamid))
	b := &Beam{
		ship.status.X, ship.status.Y, ship.status.Angle + 90,
		(ship.status.Angle + 90) * math.Pi / 180,
		0, beamid, nil,
	}

	b.pos = ship.beamList.PushBack(b)
	// broad to all players
	data := &ProtoShootBeam{
		BeamId: beamid, 
	}
	data.ShipStatus = ship.status
	ship.scene.BroadProto(ship, true, kCmdShootBeam, "data", data)
	return nil
}

func InitShip(ship *Ship, p Player2) {
	ship.beamList = list.New()
	InitList(&ship.sceneList, ship)
	InitList(&ship.stateList, ship)
	ship.Player2 = p
}
