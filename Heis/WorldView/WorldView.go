package WorldView

// ALT HER ER KOPIERT FRA CLAUDE

import "heis/order"

type Dirn int
const (
    D_Down Dirn = -1
    D_Stop Dirn = 0
    D_Up   Dirn = 1
)

type Behaviour int
const (
    B_Idle     Behaviour = 0
    B_Moving   Behaviour = 1
    B_DoorOpen Behaviour = 2
)

type ElevatorState struct {
    Floor     int
    Direction Dirn
    Behaviour Behaviour
    CabOrders []order.Order
}

type WorldView struct {
    HallOrders [][2]order.Order
    States     map[string]ElevatorState
}

func InitWorldView(id string, numFloors int) WorldView {
    wv := WorldView{
        HallOrders: order.InitHallOrders(numFloors),
        States:     make(map[string]ElevatorState),
    }
    wv.States[id] = ElevatorState{
        Floor:     -1,
        Direction: D_Stop,
        Behaviour: B_Idle,
        CabOrders: order.InitCabOrders(numFloors),
    }
    return wv
}