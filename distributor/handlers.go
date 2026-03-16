package distributor

import (
	"heis/WorldView"
	"heis/driver/elevio"
	"heis/order"
	"heis/saveIfCrash"
)

func handleButtonPress(id string, wv WorldView.WorldView, btn elevio.ButtonEvent, peersAlive []string) WorldView.WorldView {
	alone := isSoleElevator(peersAlive)

	switch btn.Button {
	case elevio.BT_Cab:
		state := wv.States[id]
		if state.CabOrders[btn.Floor].Status == order.OS_None ||
			state.CabOrders[btn.Floor].Status == order.OS_Unk {
			// cab orders are private
			state.CabOrders[btn.Floor].Status = order.OS_Const
			state.CabOrders[btn.Floor].Barrier = []string{}
			wv.States[id] = state
			saveIfCrash.SaveCabOrders(id, state.CabOrders)
		}
	default:
		b := int(btn.Button)
		if wv.HallOrders[btn.Floor][b].Status == order.OS_None ||
			wv.HallOrders[btn.Floor][b].Status == order.OS_Unk {
			wv.HallOrders[btn.Floor][b].Status = order.OS_Unconst
			wv.HallOrders[btn.Floor][b].Barrier = []string{id}
			if alone {
				wv.HallOrders[btn.Floor][b].Status = order.OS_Const
				wv.HallOrders[btn.Floor][b].Barrier = []string{}
			}
		}
	}
	return wv
}

func handleFinishedOrder(id string, wv WorldView.WorldView, btn elevio.ButtonEvent, peersAlive []string) WorldView.WorldView {
	alone := isSoleElevator(peersAlive)

	switch btn.Button {
	case elevio.BT_Cab:
		state := wv.States[id]
		state.CabOrders[btn.Floor].Status = order.OS_None
		state.CabOrders[btn.Floor].Barrier = []string{}
		wv.States[id] = state
		saveIfCrash.SaveCabOrders(id, state.CabOrders)
	default:
		b := int(btn.Button)
		wv.HallOrders[btn.Floor][b].Status = order.OS_Fin
		wv.HallOrders[btn.Floor][b].Barrier = []string{id}
		if alone {
			wv.HallOrders[btn.Floor][b].Status = order.OS_None
			wv.HallOrders[btn.Floor][b].Barrier = []string{}
		}
	}
	return wv
}

func isSoleElevator(peersAlive []string) bool {
	const minimumPeersForBarrier = 2
	return len(peersAlive) < minimumPeersForBarrier
}
