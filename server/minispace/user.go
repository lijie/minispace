package minispace

import _ "code.google.com/p/go.net/websocket"
import "fmt"
import "time"
import "sync"
import "math"

type UserDb struct {
	Name string `bson:"_id"`
	Pass string
	LoginTime int64
	RegTime int64
	BestScore int
	Win int
	Lose int
}

type ShipAttr struct {
	lv, speed, hp, beam int
}

type User struct {
	userdb UserDb
	Ship
	rotatedt float64
	movedt float64
	enable bool
	login bool
	dirty bool
	eventch chan *Event
	msgch chan *Msg
	lasterr int
	conn *Client
}

func (u *User) UserEventRoutine() {
	var event *Event
	var msg *Msg

	for u.enable {
		select {
		case event = <- u.eventch:
			// be kicked
			if event.cmd == kEventKickClient {
				fmt.Printf("%s be kicked by %s\n",
					u.userdb.Name, event.sender.Name())
				// del current player from scene
				u.scene.DelPlayer(&u.Ship)
				u.enable = false
				u.Logout()
				if event.callback != nil {
					event.callback(event, nil)
				}
			}
		case msg = <- u.msgch:
			u.conn.Reply(msg)
		}
	}

	fmt.Printf("user %s end event loop\n", u.Name())
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
		sender: &u.Ship,
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

func (u *User) Logout() {
	DeleteOnline(u.userdb.Name)

	if u.dirty {
		err := SharedDB().SyncSave(u.userdb.Name, &u.userdb)
		if err != nil {
			fmt.Printf("user %s, save db failed\n", u.userdb.Name)
			return
		}

		u.dirty = false
	}
}

func init() {
	ClientProcRegister(kCmdUserUpdate, procUserUpdate)
	ClientProcRegister(kCmdUserLogin, procUserLogin)
	ClientProcRegister(kCmdUserAction, procUserAction)
//	ClientProcRegister(kCmdShipRestart, procShipRestart)
}

func procUserUpdate(user *User, msg *Msg) int {
	user.status.X = msg.Body["x"].(float64)
	user.status.Y = msg.Body["y"].(float64)
	user.status.DestX = msg.Body["destx"].(float64)
	user.status.DestY = msg.Body["desty"].(float64)
	user.status.Angle = msg.Body["angle"].(float64)
	user.status.Move = int(msg.Body["move"].(float64))
	//user.status.Rotate = int(msg.Body["rotate"].(float64))
	user.status.Rotate = ROTATE_LEFT
	user.updated = true
	fmt.Printf("%#v\n", user.status)
	return PROC_OK
}

func loadUser(user *User, userid string, password string) int {
	if miniConfig.EnableDB == false {
		user.userdb.Name = userid
		return PROC_OK
	}

	newbie := false
	fmt.Printf("start load db\n")
	err := SharedDB().SyncLoad(userid, &user.userdb)
	if err == ErrUserNotFound {
		// new user
		newbie = true
	} else if err != nil {
		user.SetErrCode(ErrCodeDBError)
		return PROC_ERR
	}
	fmt.Printf("load db done\n")

	now := time.Now()
	if !newbie {
		// check password
		fmt.Printf("registed user %#v\n", user.userdb)
		user.userdb.LoginTime = now.Unix()
		user.MarkDirty()

		// TODO: use md5 at least...
		if password != user.userdb.Pass {
			fmt.Printf("user %s password error\n", user.userdb.Name)
			return PROC_ERR
		}
	} else {
		user.userdb.Name = userid
		user.userdb.Pass = password
		user.userdb.RegTime = now.Unix()
		user.userdb.LoginTime = now.Unix()
		fmt.Printf("create new user %#v\n", user.userdb)

		// flush to db
		err = SharedDB().SyncSave(userid, &user.userdb)
		if err != nil {
			fmt.Printf("User %s save db failed\n", user.userdb.Name)
			user.SetErrCode(ErrCodeDBError)
			return PROC_ERR
		}
	}

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
		fmt.Printf("kick done\n")
	}

	password, ok := msg.Body["password"]
	if !ok || len(password.(string)) < 3 {
		fmt.Printf("client %#v no password or passwoard too short, fail\n", user)
		user.SetErrCode(ErrCodeInvalidProto)
		return PROC_ERR
	}

	ret := loadUser(user, msg.Userid, password.(string))
	if ret != PROC_OK {
		return ret
	}

	if InsertOnline(msg.Userid, user) != nil {
		return PROC_KICK
	}

	user.login = true
	fmt.Printf("try add %s to scene\n", msg.Userid)

	// ok, login succ, add to a scene
	_, err := CurrentScene().AddPlayer(&user.Ship)
	if err != nil {
		fmt.Printf("add client err")
		return PROC_KICK
	}

	// all done, send reply
	reply := NewMsg()
	reply.Cmd = kCmdUserLogin
	reply.Body["id"] = user.status.Id
	user.conn.Reply(reply)

	fmt.Printf("login result: %#v\n", reply)
	return PROC_OK
}

func procUserAction(user *User, msg *Msg) int {
	user.status.X = msg.Body["x"].(float64)
	user.status.Y = msg.Body["y"].(float64)
	user.status.Angle = msg.Body["angle"].(float64)
	user.status.Move = int(msg.Body["move"].(float64))
	user.status.Rotate = int(msg.Body["rotate"].(float64))

	act := int(msg.Body["act"].(float64))
	if act == 1 {
		beamid := int(msg.Body["beamid"].(float64))
		err := user.AddBeam(beamid)
		if err != nil {
			return PROC_ERR
		}
	}

	return PROC_OK
}

func procShipRestart(user *User, msg *Msg) int {
	return PROC_OK
}

// for Player interface

func (user *User) SendMsg(msg *Msg) error {
	select {
	case user.msgch <- msg:
		break
	default:
		fmt.Printf("send to client %s but channel is full\n", user.userdb.Name)
	}

	return nil
}

func (user *User) Name() string {
	return user.userdb.Name
}

func (user *User) calcDestRotate() {
	st := &user.status
	r := math.Atan((st.DestY - st.Y) / (st.DestX - st.X));

	r = r / math.Pi * 180;
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

	r4 := r3 - r2;
	if r4 < 0 {
		r4 = r4 + 360
	}

	// move_ = MOVE_FORWARD;
	if r4 < 180 {
		st.Rotate = ROTATE_RIGHT;
		user.rotatedt = r4 / 120;
	} else {
		st.Rotate = ROTATE_LEFT;
		user.rotatedt = (360 - r4) / 120;
	}
}

func (user *User) calcDestMove() {
	st := &user.status
	y := st.DestY - st.Y
	x := st.DestX - st.X
	user.movedt = math.Sqrt(y*y + x*x) / kShipSpeed
	// fmt.Printf("movedt %f", user.movedt)
}

func (user *User) doRotate(dt float64) {
	st := &user.status
	if st.Move == MOVE_NONE || st.Rotate == ROTATE_NONE {
		return
	}

	user.calcDestRotate()
	if user.rotatedt <= 0 {
		user.rotatedt = 0
		st.Rotate = ROTATE_NONE
		fmt.Printf("should stop rotate2\n")
		return
	}

	if dt > user.rotatedt {
		dt = user.rotatedt
		user.rotatedt = 0
		fmt.Printf("should stop rotate\n")
	} else {
		user.rotatedt -= dt
	}

	var angle float64
	if st.Rotate == ROTATE_LEFT {
		angle = st.Angle - 120 * dt
	} else if st.Rotate == ROTATE_RIGHT {
		angle = st.Angle + 120 * dt
	} else {
		return
	}

	if angle < 0 {
		angle = angle + 360
	} else if angle >= 360 {
		angle = angle - 360
	}

	st.Angle = angle
	if user.rotatedt == 0 {
		st.Rotate = ROTATE_NONE
	}
}

func (user *User) checkMove() {
	st := &user.status
	y := st.DestY - st.Y
	x := st.DestX - st.X
	dist := math.Sqrt(y*y + x*x)

	if dist > 10 {
		fmt.Printf("WARNING: invalid move!\n")
	}
}

func (user *User) doMove(dt float64) {
	st := &user.status
	if st.Move == MOVE_NONE {
		return
	}

	user.calcDestMove()
	if user.movedt <= 0 {
		fmt.Printf("ship %d is stop\n", st.Id)
		user.checkMove()
		st.Move = MOVE_NONE
		st.X = st.DestX
		st.Y = st.DestY
		return
	}

	user.movedt -= dt
	if user.movedt < 0 {
		fmt.Printf("ship %d is stop\n", st.Id)
		user.checkMove()
		st.Move = MOVE_NONE
		st.X = st.DestX
		st.Y = st.DestY
		return
	}

	angle := st.Angle + 90
	r := kShipSpeed * dt
	x := r * math.Sin(angle / 180 * math.Pi)
	y := r * math.Cos(angle / 180 * math.Pi)

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
 
func (u *User) Update(delta float64) {
	// ms convert to second
	u.doRotate(delta / 1000)
	u.doMove(delta / 1000)

	u.sendPath()
}

func (u *User) sendPath() {
	msg := NewMsg()
	msg.Cmd = kCmdShowPath

	msg.Body["data"] = &u.status
	u.SendMsg(msg)
}

func (u *User) Dead() {
	u.userdb.Lose++
}

func (u *User) Win() {
	u.userdb.Win++
}

func InitUser(u *User, c *Client) {
	InitShip(&u.Ship, u)
	u.enable = true
	u.conn = c
	u.status.Hp = 100
	u.eventch = make(chan *Event, 8)
	u.msgch = make(chan *Msg, 64)
}
