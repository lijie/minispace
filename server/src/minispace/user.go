package minispace

import _ "code.google.com/p/go.net/websocket"
import "container/list"
import "fmt"
import "math"
import "time"
import "sync"

type UserDb struct {
	Name string `bson:"_id"`
	Pass string
	LoginTime int64
	RegTime int64
	BestScore int
	Win int
	Lose int
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
	enable bool
	login bool
	dirty bool
	scene *Scene
	sceneList List
	beamMap int
	beamList *list.List
	eventch chan *Event
	lasterr int
	conn *Client
}

func (u *User) UserEventRoutine() {
	var event *Event

	for u.enable {
		event = <- u.eventch

		// be kicked
		if event.cmd == kEventKickClient {
			// del current player from scene
			u.scene.DelPlayer(u)
			u.enable = false
			u.Logout()
			if event.callback != nil {
				event.callback(event, nil)
			}
		}
	}
}

func (u *User) KickName(name string) {
	other := SearchOnline(name)
	if other != nil {
		u.KickPlayer(other)
	}
}

func (u *User) KickPlayer(other *User) {
	cmd := &Event{
		cmd: kEventKickClient,
		sender: u,
	}

	var lock sync.Mutex
	lock.Lock()

	cmd.callback = func(cmd *Event, err error) {
		lock.Unlock()
	}

	// send cmd
	other.eventch <- cmd

	// wait
	lock.Lock()
}

func (u *User) SetErrCode(code int) {
	u.lasterr = code
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
			s.broadStopBeam(u, int(beam.id), 0)
		}

		b = tmp
	}
}

func (u *User) CheckHitAll(l *List, s *Scene) {
	var tmp *List

	p := l.Next()
	for p != l {
		tmp = p.Next()
		u.CheckHit(p.Host().(*User), s)
		p = tmp
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
		s.broadStopBeam(u, int(beam.id), 1)

		target.Hp -= 20
		if target.Hp < 0 {
			target.Hp = 0
		}
		if target.Hp == 0 {
			s.broadShipDead(target.Id)
			u.Win++
			target.Lose++

			// remove target from activeList
			target.sceneList.RemoveSelf()

			// add target to deadlist
			s.deadList.PushBack(&target.sceneList)
		}
		return
	}

	return
}

func (u *User) Logout() {
	DeleteOnline(u.Name)

	if u.dirty {
		err := SharedDB().SyncSave(u.Name, &u.UserDb)
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

func procUserUpdate(user *User, msg *Msg) int {
	user.X = msg.Body["x"].(float64)
	user.Y = msg.Body["y"].(float64)
	user.Angle = msg.Body["angle"].(float64)
	user.Move = int(msg.Body["move"].(float64))
	user.Rotate = int(msg.Body["rotate"].(float64))
	return PROC_OK
}

func procUserLogin(user *User, msg *Msg) int {
	if user.login {
		// repeat login request is error
		return PROC_KICK
	}

	// user already login? kick old
	if o := SearchOnline(msg.Userid); o != nil {
		fmt.Printf("kick %s for relogin\n", msg.Userid)
		user.KickPlayer(o)
	}

	password, ok := msg.Body["password"]
	if !ok || len(password.(string)) < 3 {
		fmt.Printf("client %#v no password or passwoard too short, fail\n", user)
		user.SetErrCode(ErrCodeInvalidProto)
		return PROC_ERR
	}

	newbie := false
	err := SharedDB().SyncLoad(msg.Userid, &user.UserDb)
	if err == ErrUserNotFound {
		// new user
		newbie = true
	} else if err != nil {
		user.SetErrCode(ErrCodeDBError)
		return PROC_ERR
	}

	now := time.Now()
	if !newbie {
		// check password
		fmt.Printf("registed user %#v\n", user.UserDb)
		user.UserDb.LoginTime = now.Unix()
		user.MarkDirty()

		// TODO: use md5 at least...
		if password.(string) != user.Pass {
			fmt.Printf("user %s password error\n", user.Name)
			return PROC_ERR
		}
	} else {
		user.UserDb.Name = msg.Userid
		user.UserDb.Pass = password.(string)
		user.UserDb.RegTime = now.Unix()
		user.UserDb.LoginTime = now.Unix()
		fmt.Printf("create new user %#v\n", user.UserDb)

		// flush to db
		err = SharedDB().SyncSave(msg.Userid, &user.UserDb)
		if err != nil {
			fmt.Printf("User %s save db failed\n", user.Name)
			user.SetErrCode(ErrCodeDBError)
			return PROC_ERR
		}
	}

	if InsertOnline(msg.Userid, user) != nil {
		return PROC_KICK
	}

	user.login = true

	// ok, login succ, add to a scene
	_, err = CurrentScene().AddPlayer(user)
	if err != nil {
		fmt.Printf("add client err")
		return PROC_KICK
	}

	// all done, send reply
	reply := NewMsg()
	reply.Cmd = kCmdUserLogin
	reply.Body["id"] = user.Id
	user.conn.Reply(reply)

	fmt.Printf("login result: %#v\n", reply)
	return PROC_OK
}

func procUserAction(user *User, msg *Msg) int {
	user.X = msg.Body["x"].(float64)
	user.Y = msg.Body["y"].(float64)
	user.Angle = msg.Body["angle"].(float64)
	user.Move = int(msg.Body["move"].(float64))
	user.Rotate = int(msg.Body["rotate"].(float64))

	act := int(msg.Body["act"].(float64))
	if act == 1 {
		beamid := msg.Body["beamid"].(float64)
		// check beamid is valid
		id := uint(beamid)
		if ((1 << id) & user.beamMap) != 0 {
			// error
			fmt.Printf("beamid %d already used\n", id)
			return PROC_ERR
		}
		// save beamid and beam
		user.beamMap |= (1 << id)
		b := &Beam{
			user.X, user.Y, user.Angle + 90,
			(user.Angle + 90) * math.Pi / 180,
			beamid, nil,
		}
		b.pos = user.beamList.PushBack(b)
		// broad to all players
		data := &ProtoShootBeam{
			BeamId: beamid, 
		}
		data.ShipStatus = user.ShipStatus
		user.scene.BroadProto(user, true, kCmdShootBeam, "data", data)
	}

	return PROC_OK
}

func InitUser(u *User, c *Client) {
	u.beamList = list.New()
	u.conn = c
	u.Hp = 100
	u.eventch = make(chan *Event, 128)
	InitList(&u.sceneList, u)
}
