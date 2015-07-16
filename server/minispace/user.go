package minispace

import (
	_ "code.google.com/p/go.net/websocket"
	"fmt"
	_ "math"
	"sync"
	"syscall"
	"time"
)

// A UserDb is a struct for saving human data to database
type UserDb struct {
	Name      string `bson:"_id"`
	Pass      string
	LoginTime int64
	RegTime   int64
	BestScore int
	Win       int
	Lose      int
}

// A ShipAttr is ... no use yet
type ShipAttr struct {
	lv, speed, hp, beam int
}

// A User is a struct for a human player
type User struct {
	userdb UserDb
	Ship
	rotatedt   float64
	movedt     float64
	enable     bool
	login      bool
	dirty      bool
	eventch    chan *Event
	msgch      chan *Msg
	lasterr    int
	t1         int64
	timeOffset int64
	conn       *Client
}

func (u *User) UserEventRoutine() {
	var event *Event
	var msg *Msg

	for u.enable {
		select {
		case event = <-u.eventch:
			// be kicked
			if event.cmd == kEventKickClient {
				fmt.Printf("%s be kicked by %s\n",
					u.userdb.Name, event.sender.Name())
				// del current player from scene
				u.scene.DelPlayer(&u.Ship)
				u.enable = false
				u.logout()
				if event.callback != nil {
					event.callback(event, nil)
				}
			}
		case msg = <-u.msgch:
			u.conn.Reply(msg)
		}
	}

	fmt.Printf("user %s end event loop\n", u.Name())
}

// KickName kicks a player to offline by name
func (u *User) KickName(name string) {
	other := SearchOnline(name)
	if other != nil {
		u.KickPlayer(other)
	}
}

// KickPlayer kicks a player to offline
func (u *User) KickPlayer(other *User) {
	cmd := &Event{
		cmd:    kEventKickClient,
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

func (u *User) setErrCode(code int) {
	u.lasterr = code
}

func (u *User) markDirty() {
	u.dirty = true
}

func (u *User) logout() {
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
	ClientProcRegister(kCmdDetectDelay, procDetectDelay)
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
		user.setErrCode(ErrCodeDBError)
		return PROC_ERR
	}
	fmt.Printf("load db done\n")

	now := time.Now()
	if !newbie {
		// check password
		fmt.Printf("registed user %#v\n", user.userdb)
		user.userdb.LoginTime = now.Unix()
		user.markDirty()

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
			user.setErrCode(ErrCodeDBError)
			return PROC_ERR
		}
	}

	return PROC_OK
}

func procDetectDelay(user *User, msg *Msg) int {
	sec := int64(msg.Body["sec"].(float64))
	usec := int64(msg.Body["usec"].(float64))
	t2 := sec*1000000 + usec

	var tv syscall.Timeval
	syscall.Gettimeofday(&tv)
	t3 := tv.Sec*1000000 + tv.Usec

	delay := (t3 - user.t1) / 2
	user.timeOffset = t3 - delay - t2

	fmt.Printf("user %s t1 %d t2 %d t3 %d delay %d timeoffset %d\n",
		user.userdb.Name, user.t1, t2, t3, delay, user.timeOffset)
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
		user.setErrCode(ErrCodeInvalidProto)
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
		err, _ := user.AddBeam(beamid)
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
	// u.sendPath()
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
