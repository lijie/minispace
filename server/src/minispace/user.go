package minispace

import _ "code.google.com/p/go.net/websocket"
import "fmt"
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

type ShipAttr struct {
	lv, speed, hp, beam int
}

type User struct {
	userdb UserDb
	Ship
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
				fmt.Printf("%s be kicked by %s\n", u.UserName(), event.sender.UserName())
				// del current player from scene
				u.scene.DelPlayer(u)
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

	fmt.Printf("user %s end event loop\n", u.UserName())
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

func (u *User) Update(delta float64) {
//	var tmp *list.Element
//	var beam *Beam
//
//	for b := u.beamList.Front(); b != nil; {
//		beam = b.Value.(*Beam)
//		tmp = b.Next()
//
//		if !beam.Update(delta) {
//			u.beamList.Remove(b)
//			u.beamMap = u.beamMap &^ (1 << uint(beam.id))
//			u.scene.broadStopBeam(u, int(beam.id), 0)
//		}
//
//		b = tmp
//	}
	ShipUpdateBeam(u, delta)
}

func (u *User) GetShip() *Ship {
	return &u.Ship
}

func (u *User) Beat() {
	u.userdb.Win++
}

func (u *User) SetDead() {
	u.Ship.SetDead()
	u.userdb.Lose++
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
	user.status.Angle = msg.Body["angle"].(float64)
	user.status.Move = int(msg.Body["move"].(float64))
	user.status.Rotate = int(msg.Body["rotate"].(float64))
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

	newbie := false
	fmt.Printf("start load db\n")
	err := SharedDB().SyncLoad(msg.Userid, &user.userdb)
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
		if password.(string) != user.userdb.Pass {
			fmt.Printf("user %s password error\n", user.userdb.Name)
			return PROC_ERR
		}
	} else {
		user.userdb.Name = msg.Userid
		user.userdb.Pass = password.(string)
		user.userdb.RegTime = now.Unix()
		user.userdb.LoginTime = now.Unix()
		fmt.Printf("create new user %#v\n", user.userdb)

		// flush to db
		err = SharedDB().SyncSave(msg.Userid, &user.userdb)
		if err != nil {
			fmt.Printf("User %s save db failed\n", user.userdb.Name)
			user.SetErrCode(ErrCodeDBError)
			return PROC_ERR
		}
	}

	if InsertOnline(msg.Userid, user) != nil {
		return PROC_KICK
	}

	user.login = true
	fmt.Printf("try add %s to scene\n", user.userdb.Name)

	// ok, login succ, add to a scene
	_, err = CurrentScene().AddPlayer(user)
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
		err := ShipAddBeam(user, beamid)
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

func (user *User) SendClient(msg *Msg) error {
	select {
	case user.msgch <- msg:
		break
	default:
		fmt.Printf("send to client %s but channel is full\n", user.userdb.Name)
	}

	return nil
}

func (user *User) UserName() string {
	return user.userdb.Name
}

func InitUser(u *User, c *Client) {
	InitShip(&u.Ship)
	u.enable = true
	u.conn = c
	u.status.Hp = 100
	u.eventch = make(chan *Event, 8)
	u.msgch = make(chan *Msg, 64)
	InitList(&u.sceneList, u)
	InitList(&u.stateList, u)
}
