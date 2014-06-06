package minispace

import "errors"

// cmd
const (
	kCmdUserLogin = 1
	kCmdUserUpdate = 2
	kCmdUserNotify = 3
	kCmdUserKick = 4
	kCmdUserAction = 5
)

const (
	kActionShoot = 1
)

// error
var ErrReadPacket = errors.New("Read packet from client error")
var ErrSceneFull = errors.New("Scene is full")
