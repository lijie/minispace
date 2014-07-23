package minispace

import _ "code.google.com/p/go.net/websocket"
import "fmt"
import "time"
import "sync"
import _ "math"

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
	ClientProcRegister(kCmdSetTarget, procSetTarget)
//	ClientProcRegister(kCmdShipRestart, procShipRestart)
}

func procSetTarget(user *User, msg *Msg) int {
	user.status.X = msg.Body["x"].(float64)
	user.status.Y = msg.Body["y"].(float64)
	user.status.Angle = msg.Body["angle"].(float64)
	user.status.Move = int(msg.Body["move"].(float64))
	user.status.Rotate = int(msg.Body["rotate"].(float64))
	targetid := msg.Body["targetid"].(float64)
	user.updated = true

	user.SetTarget(int(targetid))
	return PROC_OK
}

func procUserUpdate(user *User, msg *Msg) int {
	user.status.X = msg.Body["x"].(float64)
	user.status.Y = msg.Body["y"].(float64)
	user.status.DestX = msg.Body["destx"].(float64)
	user.status.DestY = msg.Body["desty"].(float64)
	user.status.Angle = msg.Body["angle"].(float64)
	user.status.Move = int(msg.Body["move"].(float64))
	user.status.Rotate = int(msg.Body["rotate"].(float64))
	user.updated = true

	// if has new dest XY, clear target
	if user.status.Move == MOVE_FORWARD {
		user.target = nil
	}

	// user.scene.StartRunFrame()
	// ship.Update() will simulate ship move and rotate
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

func (u *User) Update(delta float64) {
	// ms convert to second
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
