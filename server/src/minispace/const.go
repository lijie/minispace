package minispace

import "errors"

// cmd
const (
	kCmdUserLogin = 1
	kCmdUserUpdate = 2
	kCmdShipStatus = 3
	kCmdUserKick = 4
	kCmdUserAction = 5
	kCmdAddUser = 6
	kCmdStopBeam = 7
	kCmdShootBeam = 8
	kCmdShipDead = 9
)

const (
	kActionShoot = 1
)

// internal event
const (
	kEventAddClient = iota
	kEventDelClient
	kEventKickClient
	kEventDBLoad
	kEventDBSave
)

// screen size
const (
	kScreenWidth = 960
	kScreenHeight = 640
)

// error
var ErrReadPacket = errors.New("Read packet from client error")
var ErrSceneFull = errors.New("Scene is full")
var ErrDBConnectFail = errors.New("Connect Database failed")
var ErrFifoIsFull = errors.New("Fifo is full")
var ErrFifoIsEmpty = errors.New("Fifo is empty")
var ErrFifoInvalidSize = errors.New("Invalid fifo size")
var ErrUserNotFound = errors.New("User not found")
var ErrUserAlreadyLogin = errors.New("User already login")

// error code
const (
	ErrCodeInvalidProto = 0x1000
	ErrCodeDBError = 0x1001
)
