package distributor

import (
	"heis/config"
	"heis/driver/elevio"
	"heis/WorldView"
	"heis/order"
    "heis/elevtype"

)

func updateButtonLamps(id string, wv  worldview.WorldView) {
    for f := 0; f < config.N_FLOORS; f++ {
        // Hall up lamp
        elevio.SetButtonLamp(elevio.BT_HallUp, f,
            wv.HallOrders[f][int(elevtype.B_HallUp)].Status >= order.OS_Const)
        // Hall down lamp
        elevio.SetButtonLamp(elevio.BT_HallDown, f,
            wv.HallOrders[f][int(elevtype.B_HallDown)].Status >= order.OS_Const)
    }
    // Cab lamps — only for this elevator's own cab orders
    if state, ok := wv.States[id]; ok {
        for f := 0; f < config.N_FLOORS; f++ {
            elevio.SetButtonLamp(elevio.BT_Cab, f,
                state.CabOrders[f].Status >= order.OS_Const)
        }
    }
}