package elevio

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const _pollRate = 20 * time.Millisecond

const (
	driverMessageSize = 4
	defaultNumFloors  = 4
	numButtonTypes    = 3
	invalidFloor      = -1
)

const (
	cmdSetMotorDirection byte = 1
	cmdSetButtonLamp     byte = 2
	cmdSetFloorIndicator byte = 3
	cmdSetDoorOpenLamp   byte = 4
	cmdSetStopLamp       byte = 5
	cmdGetButton         byte = 6
	cmdGetFloor          byte = 7
	cmdGetStop           byte = 8
	cmdGetObstruction    byte = 9
)

const (
	responseStatusIndex = 1
	responseFloorIndex  = 2
)

var _initialized bool = false
var _numFloors   int  = defaultNumFloors
var _mtx         sync.Mutex
var _conn        net.Conn

type MotorDirection int

const (
	MotorDirectionUp   MotorDirection = 1
	MotorDirectionDown                = -1
	MotorDirectionStop                = 0
)

type ButtonType int

const (
	ButtonTypeHallUp   ButtonType = 0
	ButtonTypeHallDown            = 1
	ButtonTypeCab                 = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

func Init(addr string, numFloors int) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_numFloors = numFloors
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
}

func SetMotorDirection(dir MotorDirection) {
	write([driverMessageSize]byte{cmdSetMotorDirection, byte(dir), 0, 0})
}

func SetButtonLamp(button ButtonType, floor int, value bool) {
	write([driverMessageSize]byte{cmdSetButtonLamp, byte(button), byte(floor), toByte(value)})
}

func SetFloorIndicator(floor int) {
	write([driverMessageSize]byte{cmdSetFloorIndicator, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	write([driverMessageSize]byte{cmdSetDoorOpenLamp, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	write([driverMessageSize]byte{cmdSetStopLamp, toByte(value), 0, 0})
}

func PollButtons(receiver chan<- ButtonEvent) {
	prev := make([][numButtonTypes]bool, _numFloors)
	for {
		time.Sleep(_pollRate)
		for f := 0; f < _numFloors; f++ {
			for b := ButtonType(0); b < ButtonType(numButtonTypes); b++ {
				v := GetButton(b, f)
				if v != prev[f][b] && v != false {
					receiver <- ButtonEvent{Floor: f, Button: ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

func PollFloorSensor(receiver chan<- int) {
	prev := invalidFloor
	for {
		time.Sleep(_pollRate)
		v := GetFloor()
		if v != prev && v != invalidFloor {
			receiver <- v
		}
		prev = v
	}
}

func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func GetButton(button ButtonType, floor int) bool {
	a := read([driverMessageSize]byte{cmdGetButton, byte(button), byte(floor), 0})
	return toBool(a[responseStatusIndex])
}

func GetFloor() int {
	a := read([driverMessageSize]byte{cmdGetFloor, 0, 0, 0})
	if a[responseStatusIndex] != 0 {
		return int(a[responseFloorIndex])
	}
	return invalidFloor
}

func GetStop() bool {
	a := read([driverMessageSize]byte{cmdGetStop, 0, 0, 0})
	return toBool(a[responseStatusIndex])
}

func GetObstruction() bool {
	a := read([driverMessageSize]byte{cmdGetObstruction, 0, 0, 0})
	return toBool(a[responseStatusIndex])
}

func read(in [driverMessageSize]byte) [driverMessageSize]byte {
	_mtx.Lock()
	defer _mtx.Unlock()
	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
	var out [driverMessageSize]byte
	_, err = _conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
	return out
}

func write(in [driverMessageSize]byte) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}