package minispace

import "fmt"
import "math"
import "container/list"

// too complicated....
type Player interface {
	SendClient(msg *Msg) error
	SetUserId(id int)
	SetScene(s *Scene)
	UserId() int
	UserName() string
	Status() *ShipStatus
	Update(delta float64)
	HpDown(value int) int
	SetDead()
	SetActive()
	Beat()
	// TODO: any better idea?
	GetShip() *Ship
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
	ShipStatus
	// beam manager
	beamMap int
	beamList *list.List
	// which scene we belong to
	scene *Scene
	sceneList List
	statusList List
	deadCD float64
	status int
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

	// if out of screen?
	if b.X < 0 || b.X > kMapWidth {
		return false
	}
	if b.Y < 0 || b.Y > kMapHeight {
		return false
	}
	return true
}

func (p *Ship) SetActive() {
	p.status = kStatusActive
}

func (p *Ship) SetDead() {
	p.status = kStatusDead
}

func (p *Ship) SetScene(s *Scene) {
	p.scene = s
}

func (p *Ship) SetUserId(id int) {
	p.Id = id
}

func (p *Ship) UserId() int {
	return p.Id
}

func (p *Ship) HpDown(value int) int {
	p.Hp -= float64(value)
	if p.Hp < 0 {
		p.Hp = 0
	}
	return int(p.Hp)
}

func (p *Ship) Status() *ShipStatus {
	return &p.ShipStatus
}

func ShipCheckHit(sender Player, target Player) bool {
	p := sender.GetShip()
	t := target.GetShip()

	if p.Id == target.UserId() {
		return false
	}

	var beam *Beam
	for b := p.beamList.Front(); b != nil; b = b.Next() {
		beam = b.Value.(*Beam)
		// TODO use float64 defalut
		if !beam.Hit(int(t.X), int(t.Y)) {
			continue
		}

		fmt.Printf("%d hit target %d\n", p.Id, t.Id)

		p.beamList.Remove(b)
		p.beamMap = p.beamMap &^ (1 << uint(beam.id))
		p.scene.broadStopBeam(sender, int(beam.id), 1)

		// BUG:
		return true
	}

	return false
}

func ShipUpdateBeam(player Player, dt float64) {
	p := player.GetShip()

	var tmp *list.Element
	var beam *Beam

	for b := p.beamList.Front(); b != nil; {
		beam = b.Value.(*Beam)
		tmp = b.Next()

		if !beam.Update(dt) {
			p.beamList.Remove(b)
			p.beamMap = p.beamMap &^ (1 << uint(beam.id))
			p.scene.broadStopBeam(player, int(beam.id), 0)
		}

		b = tmp
	}
}

func ShipAddBeam(player Player, beamid int) error {
	ship := player.GetShip()

	// check beamid is valid
	if ((1 << uint(beamid)) & ship.beamMap) != 0 {
		// error
		fmt.Printf("beamid %d already used\n", beamid)
		return ErrInvalidBeamID
	}

	// save beamid and beam
	ship.beamMap |= (1 << uint(beamid))
	b := &Beam{
		ship.X, ship.Y, ship.Angle + 90,
		(ship.Angle + 90) * math.Pi / 180,
		beamid, nil,
	}

	b.pos = ship.beamList.PushBack(b)
	// broad to all players
	data := &ProtoShootBeam{
		BeamId: beamid, 
	}
	data.ShipStatus = ship.ShipStatus
	ship.scene.BroadProto(player, true, kCmdShootBeam, "data", data)
	return nil
}

func InitShip(p *Ship) {
	p.beamList = list.New()
}
