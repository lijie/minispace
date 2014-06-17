package minispace

import _ "code.google.com/p/go.net/websocket"
import "container/list"
import "fmt"
import "math"
import "time"

type UserDb struct {
	Name string `bson:"_id"`
	Pass string
	LoginTime int64
	RegTime int64
	BestScore int
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
	Hp int `json:"hp"`
	// ship id
	Id int `json:"id"`
}

type ShipAttr struct {
	lv, speed, hp, beam int
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
	id float64
	pos *list.Element
}

func (b *Beam) Hit(target *User) bool {
	r := Rect{int(target.X - 25), int(target.Y) - 25, 50, 50}
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
	if b.X < 0 || b.X > kScreenWidth {
		return false
	}
	if b.Y < 0 || b.Y > kScreenHeight {
		return false
	}
	return true
}

type User struct {
	UserDb
	ShipStatus
	ShipAttr
	dirty bool
	beamMap int
	beamList *list.List
	conn *Client
}

func (u *User) MarkDirty() {
	u.dirty = true
}

func (u *User) Update(delta float64, s *Scene) {
	var tmp *list.Element
	var beam *Beam

	for b := u.beamList.Front(); b != nil; {
		beam = b.Value.(*Beam)
		tmp = b.Next()

		if !beam.Update(delta) {
			u.beamList.Remove(b)
			u.beamMap = u.beamMap &^ (1 << uint(beam.id))
			s.broadStopBeam(u.conn, int(beam.id), 0)
		}

		b = tmp
	}
}

func (u *User) CheckHitAll(l *list.List, s *Scene) {
	for p := l.Front(); p != nil; p = p.Next() {
		u.CheckHit(&p.Value.(*Client).User, s)
	}
}

func (u *User) CheckHit(target *User, s *Scene) {
	if u == target {
		return
	}

	var beam *Beam
	for b := u.beamList.Front(); b != nil; b = b.Next() {
		beam = b.Value.(*Beam)
		if !beam.Hit(target) {
			continue
		}

		fmt.Printf("%d hit target %d\n", u.Id, target.Id)

		u.beamList.Remove(b)
		u.beamMap = u.beamMap &^ (1 << uint(beam.id))
		s.broadStopBeam(u.conn, int(beam.id), 1)

		target.Hp -= 10
		if target.Hp < 0 {
			// will broad to all client int next frame
			// set 100 for test
			target.Hp = 100
		}
		return
	}

	return
}

func (u *User) Logout() {
	DeleteOnline(u.Name)

	if u.dirty {
		err := SharedDB().Save(u.Name, &u.UserDb)
		if err != nil {
			fmt.Printf("user %s, save db failed\n", u.Name)
			return
		}

		u.dirty = false
	}
}

func init() {
	ClientProcRegister(kCmdUserUpdate, procUserUpdate)
	ClientProcRegister(kCmdUserLogin, procUserLogin)
	ClientProcRegister(kCmdUserAction, procUserAction)
}

func procUserUpdate(c *Client, msg *Msg) int {
	c.X = msg.Body["x"].(float64)
	c.Y = msg.Body["y"].(float64)
	c.Angle = msg.Body["angle"].(float64)
	c.Move = int(msg.Body["move"].(float64))
	c.Rotate = int(msg.Body["rotate"].(float64))
	return PROC_OK
}

func procUserLogin(c *Client, msg *Msg) int {
	if c.login {
		// repeat login request is error
		return PROC_KICK
	}

	// user already login? kick old
	if o := SearchOnline(msg.Userid); o != nil {
		fmt.Printf("kick %s for relogin\n", msg.Userid)
		c.KickClient(o)
	}

	password, ok := msg.Body["password"]
	if !ok || len(password.(string)) < 3 {
		fmt.Printf("client %#v no password or passwoard too short, fail\n", c)
		c.SetErrCode(ErrCodeInvalidProto)
		return PROC_ERR
	}

	newbie := false
	err := SharedDB().Load(msg.Userid, &c.User.UserDb)
	if err == ErrUserNotFound {
		// new user
		newbie = true
	} else if err != nil {
		c.SetErrCode(ErrCodeDBError)
		return PROC_ERR
	}

	now := time.Now()
	if !newbie {
		// check password
		fmt.Printf("registed user %#v\n", c.User.UserDb)
		c.User.UserDb.LoginTime = now.Unix()
		c.MarkDirty()

		// TODO: use md5 at least...
		if password.(string) != c.Pass {
			fmt.Printf("user %s password error\n", c.Name)
			return PROC_ERR
		}
	} else {
		c.User.UserDb.Name = msg.Userid
		c.User.UserDb.Pass = password.(string)
		c.User.UserDb.RegTime = now.Unix()
		c.User.UserDb.LoginTime = now.Unix()
		fmt.Printf("create new user %#v\n", c.User.UserDb)

		// flush to db
		err = SharedDB().Save(msg.Userid, &c.User.UserDb)
		if err != nil {
			fmt.Printf("User %s save db failed\n", c.Name)
			c.SetErrCode(ErrCodeDBError)
			return PROC_ERR
		}
	}

	if InsertOnline(msg.Userid, c) != nil {
		return PROC_KICK
	}

	// ok, login succ, add to a scene
	_, err = CurrentScene().AddClient(c)
	if err != nil {
		fmt.Printf("add client err")
		return PROC_KICK
	}

	// all done, send reply
	c.login = true
	reply := NewMsg()
	reply.Cmd = kCmdUserLogin
	reply.Body["id"] = c.Id
	c.Reply(reply)

	fmt.Printf("login result: %#v\n", reply)
	return PROC_OK
}

type ProtoShootBeam struct {
	ShipStatus
	BeamId float64 `json:"beamid"`
}

func procUserAction(c *Client, msg *Msg) int {
	c.X = msg.Body["x"].(float64)
	c.Y = msg.Body["y"].(float64)
	c.Angle = msg.Body["angle"].(float64)
	c.Move = int(msg.Body["move"].(float64))
	c.Rotate = int(msg.Body["rotate"].(float64))

	act := int(msg.Body["act"].(float64))
	if act == 1 {
		beamid := msg.Body["beamid"].(float64)
		// check beamid is valid
		id := uint(beamid)
		if ((1 << id) & c.beamMap) != 0 {
			// error
			fmt.Printf("beamid %d already used\n", id)
			return PROC_ERR
		}
		// save beamid and beam
		c.beamMap |= (1 << id)
		b := &Beam{
			c.X, c.Y, c.Angle + 90,
			(c.Angle + 90) * math.Pi / 180,
			beamid, nil,
		}
		b.pos = c.beamList.PushBack(b)
		// broad to all players
		data := &ProtoShootBeam{
			BeamId: beamid, 
		}
		data.ShipStatus = c.User.ShipStatus
		c.scene.BroadProto(c, true, kCmdShootBeam, "data", data)
	}

	return PROC_OK
}

func InitUser(u *User, c *Client) {
	u.beamList = list.New()
	u.conn = c
	u.Hp = 100
}
