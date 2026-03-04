package main
import "time"

//from elvator.h
const N_FLOORS 	= 4
const N_BUTTONS = 3

type Dirn int
const(
	D_Down Dirn = -1
	D_Stop Dirn = 0
	D_Up Dirn 	= 1
)

type Button int 
const(
	B_HallUp Button	= iota
	B_HallDown
	B_Cab
)

type ElevatorState int
const(
	ES_Idle ElevatorState = iota
	ES_DoorOpen
	ES_Moving
)
type Config struct{
	DoorOpenDurations time.Duration
}

type Elevator struct{
	floor		int
	dirn		Dirn
	requests 	[N_FLOORS][N_BUTTONS]bool
	state 		ElevatorState
	config		Config
}
//from request.h
type DirnBehaviourPair struct{
	dirn	Dirn
	state	ElevatorState
}
