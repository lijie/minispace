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
	kCmdShipRestart = 10
)

const (
	kActionShoot = 1
)

// internal event
const (
	kEventAddPlayer = 1
	kEventDelPlayer = 2
	kEventKickClient = 3
	kEventDBLoad = 4
	kEventDBSave = 5
)

const (
	kStateActive = 0
	kStateDead = 1
)

// screen size
const (
	kScreenWidth = 960
	kScreenHeight = 640
	kMapWidth = kScreenWidth * 2
	kMapHeight = kScreenHeight * 2
)

const (
	kShipSpeed = 160
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
var ErrInvalidBeamID = errors.New("Invalid beam id")

// error code
const (
	ErrCodeInvalidProto = 0x1000
	ErrCodeDBError = 0x1001
)
