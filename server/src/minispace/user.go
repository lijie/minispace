package minispace

import _ "code.google.com/p/go.net/websocket"

type User struct {
	x, y, ro, move, rotate, act, id int
}

func init() {
	ClientProcRegister(kCmdUserUpdate, procUserUpdate)
	ClientProcRegister(kCmdUserLogin, procUserLogin)
	ClientProcRegister(kCmdUserAction, procUserAction)
}

func procUserUpdate(c *Client, msg *Msg) int {
	c.x = int(msg.Body["x"].(float64))
	c.y = int(msg.Body["y"].(float64))
	c.ro = int(msg.Body["ro"].(float64))
	c.move = int(msg.Body["move"].(float64))
	c.rotate = int(msg.Body["rotate"].(float64))
	return 0
}

func procUserLogin(c *Client, msg *Msg) int {
	c.login = true
	reply := NewMsg()
	reply.Cmd = kCmdUserLogin
	reply.Body["id"] = c.id
	c.Reply(reply)
	return 0
}

func procUserAction(c *Client, msg *Msg) int {
	c.x = int(msg.Body["x"].(float64))
	c.y = int(msg.Body["y"].(float64))
	c.ro = int(msg.Body["ro"].(float64))
	c.move = int(msg.Body["move"].(float64))
	c.rotate = int(msg.Body["rotate"].(float64))
	c.act = int(msg.Body["act"].(float64))
	return 0
}
