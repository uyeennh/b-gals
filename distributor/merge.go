package distributor

import (
	"heis/WorldView"
	"heis/order"
)

func mergeHallOrders(id string, local [][2]order.Order, received [][2]order.Order, peersAlive []string) [][2]order.Order {
	for f := range local {
		for b := range local[f] {
			local[f][b] = order.MergeOrder(id, local[f][b], received[f][b], peersAlive)
		}
	}
	return local
}

func mergeStates(
	localID string,
	localStates map[string]WorldView.ElevatorState,
	senderID string,
	receivedStates map[string]WorldView.ElevatorState,
	peersAlive []string,
) map[string]WorldView.ElevatorState {

	// Update the physical state of the elevator that sent this message
	if recvState, ok := receivedStates[senderID]; ok {
		if existing, exists := localStates[senderID]; exists {
			existing.Floor = recvState.Floor
			existing.Direction = recvState.Direction
			existing.Behaviour = recvState.Behaviour
			localStates[senderID] = existing
		} else {
			// First time we hear from this elevator — add it to our WorldView
			localStates[senderID] = recvState
		}
	}

	for recvStateID, recvState := range receivedStates {
		localState, exists := localStates[recvStateID]
		if !exists {
			continue
		}
		if recvStateID == localID {
			continue
		}
		for f := range recvState.CabOrders {
			localState.CabOrders[f] = order.MergeOrder(
				localID,
				localState.CabOrders[f],
				recvState.CabOrders[f],
				peersAlive,
			)
		}
		localStates[recvStateID] = localState
	}

	return localStates
}

func copyWorldView(wv WorldView.WorldView) WorldView.WorldView {
	copiedWV := WorldView.WorldView{
		HallOrders: make([][2]order.Order, len(wv.HallOrders)),
		States:     make(map[string]WorldView.ElevatorState),
	}
	// Deep copy hall orders
	for f := range wv.HallOrders {
		for b := range wv.HallOrders[f] {
			copiedWV.HallOrders[f][b] = order.Order{
				Status:  wv.HallOrders[f][b].Status,
				Barrier: order.CopyBarrier(wv.HallOrders[f][b].Barrier),
			}
		}
	}
	// Deep copy states
	for id, state := range wv.States {
		newCabOrders := make([]order.Order, len(state.CabOrders))
		for f, o := range state.CabOrders {
			newCabOrders[f] = order.Order{
				Status:  o.Status,
				Barrier: order.CopyBarrier(o.Barrier),
			}
		}
		copiedState := WorldView.ElevatorState{
			Floor:     state.Floor,
			Direction: state.Direction,
			Behaviour: state.Behaviour,
			CabOrders: newCabOrders,
		}
		copiedWV.States[id] = copiedState
	}
	return copiedWV
}

func copyHallOrders(orders [][2]order.Order) [][2]order.Order {
	copied := make([][2]order.Order, len(orders))
	for f := range orders {
		for b := range orders[f] {
			copied[f][b] = order.Order{
				Status:  orders[f][b].Status,
				Barrier: order.CopyBarrier(orders[f][b].Barrier),
			}
		}
	}
	return copied
}
