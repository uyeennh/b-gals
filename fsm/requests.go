package fsm
import (
	"heis/config"
	"heis/driver/elevio"

)

func ChooseDirection(e Elevator) DirnBehaviourPair {
	f := e.floor

	if f < 0 || f >= config.N_FLOORS {
		return DirnBehaviourPair{dirn: D_Stop, state: ES_Idle}
	}
	here := hasRequestsAt(e, f)
	above := requestsAbove(e, f)
	below := requestsBelow(e, f)

	switch e.dirn {
	case D_Up:
		switch {
		case above:
			return DirnBehaviourPair{dirn: D_Up, state: ES_Moving} //skal vi ha _ eller konsekvent ike spør senere
		case here:
			return DirnBehaviourPair{dirn: D_Stop, state: ES_DoorOpen} //Vurdere om ddown er bedre
		case below:
			return DirnBehaviourPair{dirn: D_Down, state: ES_Moving}
		default:
			return DirnBehaviourPair{dirn: D_Stop, state: ES_Idle}
		}

	case D_Down:
		switch {
		case below: //check below first when going down
			return DirnBehaviourPair{dirn: D_Down, state: ES_Moving}
		case here:
			return DirnBehaviourPair{dirn: D_Stop, state: ES_DoorOpen}
		case above:
			return DirnBehaviourPair{dirn: D_Up, state: ES_Moving} //skal vi ha _ eller konsekvent ike spør senere
		default:
			return DirnBehaviourPair{dirn: D_Stop, state: ES_Idle}
		}

	default: 
		//return DirnBehaviourPair{dirn: D_Stop, state: ES_Idle}

		switch {
		case here:
			return DirnBehaviourPair{dirn: D_Stop, state: ES_DoorOpen}
		case above:
			return DirnBehaviourPair{dirn: D_Up, state: ES_Moving}
		case below:
			return DirnBehaviourPair{dirn: D_Down, state: ES_Moving}
		default:
			return DirnBehaviourPair{dirn: D_Stop, state: ES_Idle}
		}

	}

}

func ShouldStop(e Elevator) bool {
	f := e.floor
	if f < 0 || f >= config.N_FLOORS {
		return false
	}

	switch e.dirn {
	case D_Down:
		if e.requests[f][B_Cab] || e.requests[f][B_HallDown] {
			return true
		}
		return e.requests[f][B_HallUp] && !requestsBelow(e, f)

	case D_Up:
		if e.requests[f][B_Cab] || e.requests[f][B_HallUp] {
			return true
		}
		return e.requests[f][B_HallDown] && !requestsAbove(e, f)

	case D_Stop:
		return hasRequestsAt(e, f)

	default:
		return false
	}
}

func ClearAtFloor(e *Elevator, floor int, finReqCh chan<- elevio.ButtonEvent) {
    if floor < 0 || floor >= config.N_FLOORS {
        return
    }

    if e.requests[floor][B_Cab] {
        finReqCh <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab}
        e.requests[floor][B_Cab] = false
    }

    switch e.dirn {
    case D_Up:
        if e.requests[floor][B_HallUp] {
            finReqCh <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallUp}
            e.requests[floor][B_HallUp] = false
        }
        if !requestsAbove(*e, floor) && e.requests[floor][B_HallDown] {
            finReqCh <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallDown}
            e.requests[floor][B_HallDown] = false
        }
    case D_Down:
        if e.requests[floor][B_HallDown] {
            finReqCh <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallDown}
            e.requests[floor][B_HallDown] = false
        }
        if !requestsBelow(*e, floor) && e.requests[floor][B_HallUp] {
            finReqCh <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallUp}
            e.requests[floor][B_HallUp] = false
        }
    case D_Stop:
        for b := Button(0); b < config.N_BUTTONS; b++ {
            if e.requests[floor][b] {
                finReqCh <- elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(b)}
                e.requests[floor][b] = false
            }
        }
    }
}

/*
func ClearAtFloor(e *Elevator, floor int, finReqCh chan<- elevio.ButtonEvent) {
	if floor < 0 || floor >= config.N_FLOORS {
		return
	}

	e.requests[floor][B_Cab] = false
	finReqCh <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab}


	switch e.dirn {
	case D_Up:
		e.requests[floor][B_HallUp] = false
		finReqCh <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallUp}
		if !requestsAbove(*e, floor) {
			e.requests[floor][B_HallDown] = false
			finReqCh <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallDown}
		}
	case D_Down:
		e.requests[floor][B_HallDown] = false
		finReqCh <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallDown}
		if !requestsBelow(*e, floor) {
			e.requests[floor][B_HallUp] = false
			finReqCh <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallUp}
		}
	case D_Stop:
		e.requests[floor][B_HallUp] = false
		finReqCh <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallUp}
		e.requests[floor][B_HallDown] = false
		finReqCh <- elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallDown}
	}
}

*/ //denne er her er ingeborg sin gamle kode. 

func hasRequestsAt(e Elevator, floor int) bool {
	for b := 0; b < config.N_BUTTONS; b++ {
		if e.requests[floor][b] {
			return true
		}
	}
	return false
}

func requestsAbove(e Elevator, floor int) bool {
	for f := floor + 1; f < config.N_FLOORS; f++ {
		if hasRequestsAt(e, f) {
			return true
		}
	}
	return false
}

func requestsBelow(e Elevator, floor int) bool {
	for f := 0; f < floor; f++ {
		if hasRequestsAt(e, f) {
			return true
		}
	}
	return false
}

func ClearNow(e Elevator, ButtonFloor int, ButtonType Button) bool {
	return e.floor == ButtonFloor &&
		((e.dirn == D_Up && ButtonType == B_HallUp) ||
			(e.dirn == D_Down && ButtonType == B_HallDown) ||
			e.dirn == D_Stop ||
			ButtonType == B_Cab)

}
