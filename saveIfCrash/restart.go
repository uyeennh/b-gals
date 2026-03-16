package saveIfCrash

import (
	"encoding/json"
	"os"
	"heis/order"
)

//saves caborders to disk

func SaveCabOrders(id string, orders []order.Order) error {
    data, err := json.Marshal(orders)
    if err != nil {
        return err
    }
    return os.WriteFile("cab_orders_"+id+".json", data, 0644)
}

//caborders read back from disk after restart

func LoadCabOrders(id string, numFloors int) []order.Order {
    data, err := os.ReadFile("cab_orders_" + id + ".json")
    if err != nil {
        return order.NewCabOrders(numFloors)
    }
    var orders []order.Order
    if err := json.Unmarshal(data, &orders); err != nil {
        return order.NewCabOrders(numFloors)
    }
    return orders
}