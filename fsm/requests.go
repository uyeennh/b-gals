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
			return DirnBehaviourPair{dirn: D_Up, state: ES_Moving} 
		case here:
			return DirnBehaviourPair{dirn: D_Stop, state: ES_DoorOpen} 
		case below:
			return DirnBehaviourPair{dirn: D_Down, state: ES_Moving}
		default:
			return DirnBehaviourPair{dirn: D_Stop, state: ES_Idle}
		}

	case D_Down:
		switch {
		case below: 
			return DirnBehaviourPair{dirn: D_Down, state: ES_Moving}
		case here:
			return DirnBehaviourPair{dirn: D_Stop, state: ES_DoorOpen}
		case above:
			return DirnBehaviourPair{dirn: D_Up, state: ES_Moving} 
		default:
			return DirnBehaviourPair{dirn: D_Stop, state: ES_Idle}
		}

	default: 

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

func collectClearedEvents(e *Elevator, floor int) ([]elevio.ButtonEvent, bool) {  
    if floor < 0 || floor >= config.N_FLOORS {
        return nil, false
    }
	var events []elevio.ButtonEvent
	needsDirectionChange := false

    if e.requests[floor][B_Cab] {
        events = append(events, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_Cab})
        e.requests[floor][B_Cab] = false
    }

    switch e.dirn {
    case D_Up:
        if e.requests[floor][B_HallUp] {
            events = append(events, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallUp})
            e.requests[floor][B_HallUp] = false
        }
        if !requestsAbove(*e, floor) && e.requests[floor][B_HallDown] {
            needsDirectionChange = true
        }
    case D_Down:
        if e.requests[floor][B_HallDown] {
            events = append(events, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallDown})
            e.requests[floor][B_HallDown] = false
        }
        if !requestsBelow(*e, floor) && e.requests[floor][B_HallUp] {
            needsDirectionChange = true
        }
    case D_Stop:
		if e.requests[floor][B_HallUp] {
			events = append(events, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallUp})
			e.requests[floor][B_HallUp] = false
			if e.requests[floor][B_HallDown] {
				needsDirectionChange = true
			}
		} else if e.requests[floor][B_HallDown] {
			events = append(events, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallDown})
			e.requests[floor][B_HallDown] = false
		}
    }
	return events, needsDirectionChange
}


func collectDirectionChangeClearedEvents(e *Elevator, floor int) []elevio.ButtonEvent {
    if floor < 0 || floor >= config.N_FLOORS {
        return nil
    }
    var events []elevio.ButtonEvent
    switch e.dirn {
    case D_Up:
        if e.requests[floor][B_HallDown] {
            events = append(events, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallDown})
            e.requests[floor][B_HallDown] = false
        }
    case D_Down:
        if e.requests[floor][B_HallUp] {
            events = append(events, elevio.ButtonEvent{Floor: floor, Button: elevio.BT_HallUp})
            e.requests[floor][B_HallUp] = false
        }
    }
    return events
}

func ClearAtFloor(e *Elevator, floor int, finReqCh chan<- elevio.ButtonEvent) {
    evs, _ := collectClearedEvents(e, floor)
    for _, ev := range evs {
        finReqCh <- ev
    }
}

func flushPendingFinReqs(pending []elevio.ButtonEvent, finReqCh chan<- elevio.ButtonEvent) {
	for _, ev := range pending {
		finReqCh <- ev
	}
}

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

func mergeRequestsInFlight(e Elevator, newAssignments [config.N_FLOORS][config.N_BUTTONS]bool) [config.N_FLOORS][config.N_BUTTONS]bool {
    merged := newAssignments
    for f := 0; f < config.N_FLOORS; f++ {
        floorIsAhead := (e.dirn == D_Up && f >= e.floor) || 
                        (e.dirn == D_Down && f <= e.floor)
        if !floorIsAhead {
            continue
        }
        if e.requests[f][B_Cab] {
            merged[f][B_Cab] = true
        }
        if e.dirn == D_Up && e.requests[f][B_HallUp] {
            merged[f][B_HallUp] = true
        }
        if e.dirn == D_Down && e.requests[f][B_HallDown] {
            merged[f][B_HallDown] = true
        }
    }
    return merged
}