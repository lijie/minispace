package minispace

import "errors"

// cmd
const (
	kCmdUserLogin = 1
)

// error
var ErrReadPacket = errors.New("Read packet from client error")
