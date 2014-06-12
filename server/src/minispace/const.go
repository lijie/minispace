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
var ErrDBConnectFail = errors.New("Connect Database failed")
var ErrFifoIsFull = errors.New("Fifo is full")
var ErrFifoIsEmpty = errors.New("Fifo is empty")
var ErrFifoInvalidSize = errors.New("Invalid fifo size")
var ErrUserNotFound = errors.New("User not found")

// error code
const (
	ErrCodeInvalidProto = 0x1000
	ErrCodeDBError = 0x1001
)
