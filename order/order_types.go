package order

type OrderStatus int

const (
	OS_Unknown nown       OrderStatus = -1 //or use iota -1??// We have not heard from the network yet, state is unknown
	OS_None          OrderStatus = 0  // No active order, meaning no button pressed. Resting state
	OS_Unconfirmed   OrderStatus = 1  // Button pressed and order exist locally, but unconfirmed request and no consistency across all nodes alive on network.
	OS_Confirmed     OrderStatus = 2  // Confirmed order and consistency across all nodes alive on the network, safe to run cost function and assign it. Turn lamp on
	OS_Finishedished OrderStatus = 3  // Assigned elevator has served the order and we wait for all elevators to confirm they know its done. Request finished
	// OS_close OrderStatus = 4 //Usikker på denne, but all elevators ack that order is served and done, can close the lights.
)

type Order struct {
	Status  OrderStatus // the current status of the order /the current stage in the lifecycle of the order
	Barrier []string    // the barrier to store elevators that has seen and ack the current status. Check if all currently alive elelevators are in the list using peersAlive later.

}
