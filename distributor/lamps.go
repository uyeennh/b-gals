package distributor

import (
	"heis/config"
	"heis/driver/elevio"
	"heis/elevtype"
	"heis/order"
	"heis/worldview"
)

func updateButtonLamps(id string, worldView worldview.WorldView) {
	for f := 0; f < config.N_FLOORS; f++ {
		// Hall up lamp
		elevio.SetButtonLamp(elevio.ButtonTypeHallUp, f,
			worldView.HallOrders[f][int(elevtype.B_HallUp)].Status >= order.OrderStatusConfirmed)
		// Hall down lamp
		elevio.SetButtonLamp(elevio.ButtonTypeHallDown, f,
			worldView.HallOrders[f][int(elevtype.B_HallDown)].Status >= order.OrderStatusConfirmed)
	}
	// Cab lamps — only for this elevator's own cab orders
	if state, ok := worldView.States[id]; ok {
		for f := 0; f < config.N_FLOORS; f++ {
			elevio.SetButtonLamp(elevio.ButtonTypeCab, f,
				state.CabOrders[f].Status >= order.OrderStatusConfirmed)
		}
	}
}
