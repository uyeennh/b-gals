package order

type OrderStatus int
const (
	OS_Unk OrderStatus = -1 // We have not heard from the network yet, state is unknown
	OS_None OrderStatus = 0 // No active order, meaning no button pressed. Resting state
	OS_Unconst OrderStatus = 1 // Button pressed and order exist locally, but unconfirmed request and no consistency across all nodes alive on network. 
	OS_Const OrderStatus = 2 // Confirmed order and consistency across all nodes alive on the network, safe to run cost function and assign it. Turn lamp on
	OS_Fin OrderStatus = 3 // Assigned elevator has served the order and we wait for all elevators to confirm they know its done. Request finished 
	// OS_close OrderStatus = 4 //Usikker på denne, but all elevators ack that order is served and done, can close the lights. 
)

type Order struct {
	Status OrderStatus // the current status of the order /the current stage in the lifecycle of the order
	Barrier []string // the barrier to store elevators that has seen and ack the current status. Check if all currently alive elelevators are in the list using peersAlive later. 

}

func InitOrder() Order{
	return Order{
		Status: OS_Unk,
		Barrier: make([]string, 0),
	}
}

func InitHallOrders(numFloors int) [][2]Order {
	orders := make([][2]Order, numFloors)

	for floor := 0; floor < numFloors; floor++ {
		// example orders[2][0] is the hall up button at floor 2. 
		orders[floor][0] = InitOrder() //index 0 = hall up button
		orders[floor][1] = InitOrder() //index 1 = hall down button
	}
	return orders
}

// Creates the cab button list 
func InitCabOrders(numFloors int) []Order {
	orders := make([]Order, numFloors)
	for floor := 0; floor < numFloors; floor++ {
		orders[floor] = InitOrder()
	}
	return orders
}



// CyclicCounter megers two version of the same request
func CyclicCounter (id string, local Order, received Order, peersAlive []string){
	//Prevent OS_None from overwriting the OS_fin 
	if local.Status == OS_None && received.Status == OS_Fin {
		return local
	}
	//  Make sure the more advanced status is adopted. 
	if local.Status < received.Status {
		return received
	}

	// If node is at OS_Fin and the received is OS_none, then the other elevator has confirmed it is done, and the node will follow them back to OS_None. 
	if local.Status == OS.Fin && received.Status == OS.None{
		return received
	}

	if local.Status == OS.Unconf || local.Status == OS_Fin{
		local.Barrier = MergeUnique(local.Barrier, received.Barrier)

	}

}