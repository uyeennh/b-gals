package worldview

import (
	"heis/elevtype"
	"heis/order"
)


// ElevatorState represents the current physical state of one elevator.
// This is what gets shared across the network so other elevators
// know where each elevator is and what it is doing.
type ElevatorState struct {
	Floor     int
	Direction elevtype.Dirn
	Behaviour elevtype.Behaviour
	CabOrders []order.Order
}

// WorldView is the full shared state of the entire elevator system.
// Every elevator keeps its own copy and they converge via broadcasting.
type WorldView struct {
	HallOrders [][2]order.Order
	States     map[string]ElevatorState
}

// InitWorldView creates a fresh WorldView for this elevator on startup.
func InitWorldView(id string, numFloors int) WorldView {
	wv := WorldView{
		HallOrders: order.NewHallOrders(numFloors),
		States:     make(map[string]ElevatorState),
	}
	wv.States[id] = ElevatorState{
		Floor:     -1,
		Direction: elevtype.D_Stop,
		Behaviour: elevtype.B_Idle,
		CabOrders: order.NewCabOrders(numFloors),
	}
	return wv
}