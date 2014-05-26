package minispace

import _ "code.google.com/p/go.net/websocket"

type User struct {
	x, y, ro, id int
}

func init() {
	ClientProcRegister(kCmdUserUpdate, procUserUpdate)
}

func procUserUpdate(c *Client, msg *Msg) int {
	c.x = int(msg.Body["x"].(float64))
	c.y = int(msg.Body["y"].(float64))
	c.ro = int(msg.Body["ro"].(float64))
	return 0
}

func procUserLogin(c *Client, msg *Msg) int {
	return 0
}
