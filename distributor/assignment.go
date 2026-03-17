package distributor

import (
	"heis/worldview"
	"heis/config"
	"heis/costfunction"
	"heis/driver/elevio"
	"heis/elevtype"
	"heis/order"
)

func computeAndSendAssignment(id string, wv  worldview.WorldView, peersAlive []string, assignmentCh chan [config.N_FLOORS][config.N_BUTTONS]bool) {
	hallReqs := extractConfirmedHallRequests(wv)
	states := buildHRAStates(wv, peersAlive)
	assigned, ok := costfunction.Compute(id, hallReqs, states)
	if !ok {
		return
	}
	reqs := mergeAssignedWithCabOrders(id, assigned, wv)

	select {
	case <-assignmentCh:
	default:
	}
	assignmentCh <- reqs

}

func extractConfirmedHallRequests(wv  worldview.WorldView) [][2]bool {
	hallReqs := make([][2]bool, config.N_FLOORS)
	for f := range wv.HallOrders {
		hallReqs[f][int(elevtype.B_HallUp)] = wv.HallOrders[f][int(elevtype.B_HallUp)].Status == order.OS_Const
		hallReqs[f][int(elevtype.B_HallDown)] = wv.HallOrders[f][int(elevtype.B_HallDown)].Status == order.OS_Const
	}
	return hallReqs
}

func buildHRAStates(wv  worldview.WorldView, peersAlive []string) map[string]costfunction.HRAElevState {
	states := make(map[string]costfunction.HRAElevState)
	for id, s := range wv.States {
		// Skip elevators with invalid floor
		if s.Floor < 0 {
			continue
		}
		// Skip elevators that are not currently alive on the network
		if !order.ContainsID(peersAlive, id) {
			continue
		}
		cabReqs := make([]bool, config.N_FLOORS)
		for f, o := range s.CabOrders {
			cabReqs[f] = o.Status == order.OS_Const
		}
		states[id] = costfunction.HRAElevState{
			Behavior:    behaviourToString(s.Behaviour),
			Floor:       s.Floor,
			Direction:   dirnToString(s.Direction),
			CabRequests: cabReqs,
		}
	}
	return states
}

func behaviourToString(b elevtype.Behaviour) string {
	switch b {
	case elevtype.B_Moving:
		return "moving"
	case elevtype.B_DoorOpen:
		return "doorOpen"
	default:
		return "idle"
	}
}

func dirnToString(d elevtype.Dirn) string {
	switch d {
	case elevtype.D_Up:
		return "up"
	case elevtype.D_Down:
		return "down"
	default:
		return "stop"
	}
}

func mergeAssignedWithCabOrders(id string, assigned [][2]bool, wv  worldview.WorldView) [config.N_FLOORS][config.N_BUTTONS]bool {
	var reqs [config.N_FLOORS][config.N_BUTTONS]bool
	state, exists := wv.States[id]
	for f := 0; f < config.N_FLOORS; f++ {
		reqs[f][int(elevio.BT_HallUp)] = assigned[f][int(elevtype.B_HallUp)]
		reqs[f][int(elevio.BT_HallDown)] = assigned[f][int(elevtype.B_HallDown)]
		if exists {
			reqs[f][int(elevio.BT_Cab)] = state.CabOrders[f].Status == order.OS_Const
		}
	}
	return reqs
}
