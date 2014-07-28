package minispace

import (
	"container/list"
	"fmt"
	"math"
)

// Player2 is an interface for all kind of player
type Player2 interface {
	SendMsg(msg *Msg) error
	Update(dt float64)
	Name() string
	Dead()
	Win()
}

// A ShipStatus saves current ship status
type ShipStatus struct {
	// position
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	DestX float64 `json:"destx"`
	DestY float64 `json:"desty"`
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

// A Ship is a base object for all player
type Ship struct {
	Player2
	status ShipStatus
	// beam manager
	beamMap  int
	beamList *list.List
	// which scene we belong to
	scene     *Scene
	sceneList List
	stateList List
	deadCD    float64
	shootcd   float64
	rotatedt  float64
	movedt    float64
	state     int
	updated   bool
	target    *Ship
}

type Rect struct {
	x, y, width, height int
}

func (r *Rect) InRect(x, y int) bool {
	if x < r.x || x > r.x+r.width {
		return false
	}
	if y < r.y || y > r.y+r.height {
		return false
	}
	return true
}

type Beam struct {
	X, Y, Angle float64
	radian      float64
	distance    float64
	id          int
	pos         *list.Element
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
	r := 1000.0 / (3.0 * 1000.0) * delta
	b.X = b.X + r*math.Sin(b.radian)
	b.Y = b.Y + r*math.Cos(b.radian)
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

// SendMsg sends msg to client
func (ship *Ship) SendMsg(msg *Msg) error {
	return ship.Player2.SendMsg(msg)
}

func (ship *Ship) doTarget() {
	if ship.target == nil {
		return
	}
	if ship.target.state == kStateDead {
		ship.target = nil
		return
	}

	st := &ship.status
	st.DestX = ship.target.status.X
	st.DestY = ship.target.status.Y
	y := st.DestY - st.Y
	x := st.DestX - st.X
	st.Move = MOVE_FORWARD
	st.Rotate = ROTATE_LEFT

	if math.Sqrt(y*y+x*x) < 100 {
		st.Move = MOVE_NONE
	}
}

func (ship *Ship) doShoot(dt float64) {
	if ship.target == nil || ship.target.state == kStateDead {
		return
	}
	ship.shootcd -= dt
	if ship.shootcd > 0 {
		return
	}
	ship.AddBeam2()
	ship.shootcd = 700
}

// Update will run for per frame
func (ship *Ship) Update(dt float64) {
	ship.updateBeam(dt)

	ship.doTarget()
	ship.doRotate(dt / 1000)
	ship.doMove(dt / 1000)
	ship.doShoot(dt)

	ship.Player2.Update(dt)
}

// Name returns player's name
func (ship *Ship) Name() string {
	return ship.Player2.Name()
}

func (p *Ship) SetActive() {
	p.state = kStateActive
}

// Dead tells player you are dead
func (p *Ship) Dead() {
	p.state = kStateDead
	p.Player2.Dead()
}

// Win tells player you win
func (p *Ship) Win() {
	p.Player2.Win()
}

// SetScene tells player whiche scene you are in
func (p *Ship) SetScene(s *Scene) {
	p.scene = s
}

// SetUserId tells player which id he owns
func (p *Ship) SetUserId(id int) {
	p.status.Id = id
}

// UserId returns player's id
func (p *Ship) UserId() int {
	return p.status.Id
}

// HpDown will be called when player is shooted
func (p *Ship) HpDown(value int) int {
	p.status.Hp -= float64(value)
	if p.status.Hp < 0 {
		p.status.Hp = 0
	}
	return int(p.status.Hp)
}

// Status returns ship status
func (p *Ship) Status() *ShipStatus {
	return &p.status
}

// ShipCheckHit checks if target be shooted by sender
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

func (ship *Ship) AddBeam2() error {
	beamid := -1
	for i := 0; i < 5; i++ {
		if ((1 << uint(i)) & ship.beamMap) == 0 {
			beamid = i
			break
		}
	}
	if beamid == -1 {
		return ErrInvalidBeamID
	}

	ship.AddBeam(beamid)
	return nil
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
	ship.scene.BroadProto(ship, false, kCmdShootBeam, "data", data)
	return nil
}

func (ship *Ship) calcDestRotate() {
	st := &ship.status
	r := math.Atan((st.DestY - st.Y) / (st.DestX - st.X))

	r = r / math.Pi * 180
	r2 := st.Angle
	var r3 float64

	if r < 0 {
		if st.DestY > st.Y {
			r3 = 360 - (180 + r)
		} else {
			r3 = math.Abs(r)
		}
	} else {
		if st.DestX > st.X {
			r3 = 360 - r
		} else {
			r3 = 180 - r
		}
	}

	r4 := r3 - r2
	if r4 < 0 {
		r4 = r4 + 360
	}

	// move_ = MOVE_FORWARD;
	if r4 < 180 {
		st.Rotate = ROTATE_RIGHT
		ship.rotatedt = r4 / 120
	} else {
		st.Rotate = ROTATE_LEFT
		ship.rotatedt = (360 - r4) / 120
	}
}

func (ship *Ship) calcDestMove() {
	st := &ship.status
	y := st.DestY - st.Y
	x := st.DestX - st.X
	ship.movedt = math.Sqrt(y*y+x*x) / kShipSpeed
	// fmt.Printf("movedt %f", ship.movedt)
}

func (ship *Ship) doRotate(dt float64) {
	st := &ship.status
	if st.Rotate == ROTATE_NONE {
		return
	}

	ship.calcDestRotate()
	if ship.rotatedt <= 0 {
		ship.rotatedt = 0
		st.Rotate = ROTATE_NONE
		// fmt.Printf("should stop rotate2\n")
		return
	}

	if dt > ship.rotatedt {
		dt = ship.rotatedt
		ship.rotatedt = 0
		// fmt.Printf("should stop rotate\n")
	} else {
		ship.rotatedt -= dt
	}

	var angle float64
	if st.Rotate == ROTATE_LEFT {
		angle = st.Angle - 120*dt
	} else if st.Rotate == ROTATE_RIGHT {
		angle = st.Angle + 120*dt
	} else {
		return
	}

	if angle < 0 {
		angle = angle + 360
	} else if angle >= 360 {
		angle = angle - 360
	}

	st.Angle = angle
	if ship.rotatedt == 0 {
		st.Rotate = ROTATE_NONE
	}
}

func (ship *Ship) checkMove() {
	st := &ship.status
	y := st.DestY - st.Y
	x := st.DestX - st.X
	dist := math.Sqrt(y*y + x*x)

	if dist > 10 {
		fmt.Printf("WARNING: invalid move!\n")
	}
}

func (ship *Ship) doMove(dt float64) {
	st := &ship.status
	if st.Move == MOVE_NONE {
		return
	}

	ship.calcDestMove()
	if ship.movedt <= 0 {
		// fmt.Printf("ship %d is stop\n", st.Id)
		ship.checkMove()
		st.Move = MOVE_NONE
		st.X = st.DestX
		st.Y = st.DestY
		return
	}

	ship.movedt -= dt
	if ship.movedt < 0 {
		// fmt.Printf("ship %d is stop\n", st.Id)
		ship.checkMove()
		st.Move = MOVE_NONE
		st.X = st.DestX
		st.Y = st.DestY
		return
	}

	angle := st.Angle + 90
	r := kShipSpeed * dt
	x := r * math.Sin(angle/180*math.Pi)
	y := r * math.Cos(angle/180*math.Pi)

	if st.Move == MOVE_FORWARD {
		x = st.X + x
		y = st.Y + y
	} else {
		x = st.X - x
		y = st.Y - y
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

	st.X = x
	st.Y = y
}

// SetTarget tells which enemy we should shoot
func (ship *Ship) SetTarget(id int) {
	s := ship.scene
	for p := s.activeList.Next(); p != &s.activeList; p = p.Next() {
		target := p.Host().(*Ship)
		if target.status.Id == id {
			ship.target = target
			fmt.Printf("user id %d get target %d\n", ship.status.Id, id)
			return
		}
	}
}

func InitShip(ship *Ship, p Player2) {
	ship.beamList = list.New()
	InitList(&ship.sceneList, ship)
	InitList(&ship.stateList, ship)
	ship.Player2 = p
	ship.updated = false
	ship.target = nil
}
