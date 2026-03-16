package order

func NewOrder() Order{
	return Order{
		Status: OS_Unk,
		Barrier: make([]string, 0),
	}
}

func NewHallOrders(numFloors int) [][2]Order { //IS IT MAGIC NUMBERS to be writing 2 and 0 and 1 here?
	orders := make([][2]Order, numFloors)

	for floor := range orders {
		// example orders[2][0] is the hall up button at floor 2. 
		orders[floor][0] = NewOrder() //index 0 = hall up button
		orders[floor][1] = NewOrder() //index 1 = hall down button
	}
	return orders
}

// Creates the cab button list 
func NewCabOrders(numFloors int) []Order {
	orders := make([]Order, numFloors)
	for floor := range orders {
		orders[floor] = NewOrder()
	}
	return orders
}