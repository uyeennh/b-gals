package distributor

import (
	"fmt"
	"heis/config"
	"heis/driver/elevio"
	"heis/network/bcast"
	"heis/network/peers"
	"heis/order"
	"heis/saveIfCrash"
	"heis/worldview"
	"time"
)

const (
	broadcastInterval  = 50 * time.Millisecond
	lampUpdateInterval = 200 * time.Millisecond
)

type networkMessage struct {
	SenderID  string
	WorldView worldview.WorldView
}

func Distributor(
	id string,
	stateCh <-chan worldview.ElevatorState,
	finReqCh <-chan elevio.ButtonEvent,
	assignmentCh chan [config.N_FLOORS][config.N_BUTTONS]bool,
	peerUpdateCh <-chan peers.PeerUpdate,
) {
	localWorldView := worldview.InitWorldView(id, config.N_FLOORS)
	savedCabOrders := saveIfCrash.LoadCabOrders(id, config.N_FLOORS)
	for f := range savedCabOrders {
		if savedCabOrders[f].Status >= order.OrderStatusUnconfirmed {
			savedCabOrders[f].Status = order.OrderStatusConfirmed
			savedCabOrders[f].Barrier = []string{}
		}
	}
	restoredState := localWorldView.States[id]
	restoredState.CabOrders = savedCabOrders
	localWorldView.States[id] = restoredState

	peersAlive := []string{id}

	messageTx := make(chan networkMessage, 1)
	messageRx := make(chan networkMessage)
	go bcast.Transmitter(config.PortDistributor, messageTx)
	go bcast.Receiver(config.PortDistributor, messageRx)

	driverButtonCh := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(driverButtonCh)

	broadcastTicker := time.NewTicker(broadcastInterval)
	defer broadcastTicker.Stop()

	lampTicker := time.NewTicker(lampUpdateInterval)
	defer lampTicker.Stop()

	// broadcastNow sends our worldview immediately without waiting
	// for the next broadcastTicker tick. Used whenever our own state
	// changes so other elevators see it before computing assignments.
	broadcastNow := func() {
		select {
		case <-messageTx:
		default:
		}
		messageTx <- networkMessage{
			SenderID:  id,
			WorldView: copyWorldView(localWorldView),
		}
	}

	for {
		select {

		case <-broadcastTicker.C:
			broadcastNow()

		case message := <-messageRx:
			if message.SenderID == id {
				continue
			}
			localWorldView.HallOrders = mergeHallOrders(
				id,
				localWorldView.HallOrders,
				message.WorldView.HallOrders,
				peersAlive,
			)
			localWorldView.States = mergeStates(
				id,
				localWorldView.States,
				message.SenderID,
				message.WorldView.States,
				peersAlive,
			)
			// Compute assignments after receiving a message from another
			// elevator. At this point our worldview contains the sender's
			// latest state (moving/idle/floor), so the cost function sees
			// a consistent picture and assigns each order to only one elevator.
			computeAndSendAssignment(id, localWorldView, peersAlive, assignmentCh)

		case buttonEvent := <-driverButtonCh:
			localWorldView = handleButtonPress(id, localWorldView, buttonEvent, peersAlive)
			// Broadcast immediately so others learn about the new order.
			// Do NOT compute assignments here — wait until we receive
			// a message back that confirms everyone has seen the order.
			broadcastNow()

		case completedOrder := <-finReqCh:
			localWorldView = handleFinishedOrder(id, localWorldView, completedOrder, peersAlive)
			localWorldView.HallOrders = mergeHallOrders(
				id,
				localWorldView.HallOrders,
				copyHallOrders(localWorldView.HallOrders),
				peersAlive,
			)
			// Broadcast immediately so others learn the order is done.
			broadcastNow()

		case state := <-stateCh:
			elevatorState := localWorldView.States[id]
			elevatorState.Floor = state.Floor
			elevatorState.Direction = state.Direction
			elevatorState.Behaviour = state.Behaviour
			localWorldView.States[id] = elevatorState
			// Broadcast immediately so other elevators see our new state
			// (moving/idle/floor) before they compute their next assignment.
			broadcastNow()

		case peerUpdate := <-peerUpdateCh:
			peersAlive = peerUpdate.Peers
			if !order.ContainsID(peersAlive, id) {
				peersAlive = append(peersAlive, id)
			}
			fmt.Println("Peers alive:", peersAlive)
			// Recompute after peer change since available elevators changed.
			computeAndSendAssignment(id, localWorldView, peersAlive, assignmentCh)

		case <-lampTicker.C:
			updateButtonLamps(id, localWorldView)
		}
	}
}
