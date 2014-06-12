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
	Act int `json:"act"`
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
	r := 1000.0 / (3.0 * 1000.0) * delta;
	b.X = b.X + r * math.Sin(b.radian)
	b.Y = b.Y + r * math.Cos(b.radian)
	fmt.Printf("beam XY: %f, %f, %f\n", b.X, b.Y, r)

	if b.X < 0 || b.X > 960 {
		return false
	}
	if b.Y < 0 || b.Y > 640 {
		return false
	}
	return true
}

type User struct {
	UserDb
	ShipStatus
	ShipAttr
	beamList *list.List
}

func (u *User) Update(delta float64) {
	var tmp *list.Element
	for b := u.beamList.Front(); b != nil; {
		tmp = b.Next()
		if !(b.Value.(*Beam).Update(delta)) {
			u.beamList.Remove(b)
		}
		b = tmp
	}
}

func (u *User) CheckHitAll(l *list.List) {
	for p := l.Front(); p != nil; p = p.Next() {
		u.CheckHit(&p.Value.(*Client).User)
	}
}

func (u *User) CheckHit(target *User) bool {
	if u == target {
		return false
	}

	for b := u.beamList.Front(); b != nil; b = b.Next() {
		if b.Value.(*Beam).Hit(target) {
			fmt.Printf("%d hit target %d\n", u.Id, target.Id)
			return true
		}
	}

	return false
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
	password, ok := msg.Body["password"]
	if !ok {
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

	if !newbie {
		// check password
	} else {
		c.User.UserDb.Name = msg.Userid
		c.User.UserDb.Pass = password.(string)
		c.User.UserDb.RegTime = time.Now().Unix()
		// flush to db
		err = SharedDB().Save(msg.Userid, &c.User.UserDb)
		if err != nil {
			c.SetErrCode(ErrCodeDBError)
			return PROC_ERR
		}
	}

	// all done, send reply
	c.login = true
	reply := NewMsg()
	reply.Cmd = kCmdUserLogin
	reply.Body["id"] = c.Id
	c.Reply(reply)
	return PROC_OK
}

func procUserAction(c *Client, msg *Msg) int {
	c.X = msg.Body["x"].(float64)
	c.Y = msg.Body["y"].(float64)
	c.Angle = msg.Body["angle"].(float64)
	c.Move = int(msg.Body["move"].(float64))
	c.Rotate = int(msg.Body["rotate"].(float64))
	c.Act = int(msg.Body["act"].(float64))

	if c.Act == 1 {
		b := &Beam{c.X, c.Y, c.Angle + 90, (c.Angle + 90) * math.Pi / 180, nil}
		b.pos = c.beamList.PushBack(b)
	}
	return PROC_OK
}

func InitUser(u *User) {
	u.beamList = list.New()
}
