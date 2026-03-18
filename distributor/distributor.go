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

	msgTx := make(chan networkMessage, 1)
	msgRx := make(chan networkMessage)
	go bcast.Transmitter(config.PortDistributor, msgTx)
	go bcast.Receiver(config.PortDistributor, msgRx)

	drvButtons := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(drvButtons)

	broadcastTicker := time.NewTicker(broadcastInterval)
	defer broadcastTicker.Stop()

	lampTicker := time.NewTicker(lampUpdateInterval)
	defer lampTicker.Stop()

	for {
		select {

		case <-broadcastTicker.C:
			msgTx <- networkMessage{
				SenderID:  id,
				WorldView: copyWorldView(localWorldView),
			}

		case msg := <-msgRx:
			if msg.SenderID == id {
				continue
			}

			localWorldView.HallOrders = mergeHallOrders(
				id,
				localWorldView.HallOrders,
				msg.WorldView.HallOrders,
				peersAlive,
			)

			localWorldView.States = mergeStates(
				id,
				localWorldView.States,
				msg.SenderID,
				msg.WorldView.States,
				peersAlive,
			)

			computeAndSendAssignment(id, localWorldView, peersAlive, assignmentCh)

		case buttonEvent := <-drvButtons:
			localWorldView = handleButtonPress(id, localWorldView, buttonEvent, peersAlive)
			computeAndSendAssignment(id, localWorldView, peersAlive, assignmentCh)
		case buttonEvent := <-finReqCh:
			localWorldView = handleFinishedOrder(id, localWorldView, buttonEvent, peersAlive)
			// Self-merge so our own barrier advances immediately
			localWorldView.HallOrders = mergeHallOrders(
				id,
				localWorldView.HallOrders,
				copyHallOrders(localWorldView.HallOrders),
				peersAlive,
			)
			computeAndSendAssignment(id, localWorldView, peersAlive, assignmentCh)

		case state := <-stateCh:
			s := localWorldView.States[id]
			s.Floor = state.Floor
			s.Direction = state.Direction
			s.Behaviour = state.Behaviour
			localWorldView.States[id] = s
			//computeAndSendAssignment(id, localWV, peersAlive, assignmentCh)

		case buttonEvent := <-peerUpdateCh:
			peersAlive = buttonEvent.Peers
			// Always make sure our own ID is in the peer list
			// peers.Receiver never includes ourselves because we never
			// hear our own broadcast — so we add ourselves manually here
			if !order.ContainsID(peersAlive, id) {
				peersAlive = append(peersAlive, id)
			}
			fmt.Println("Peers alive:", peersAlive)
			computeAndSendAssignment(id, localWorldView, peersAlive, assignmentCh)
		case <-lampTicker.C:
			updateButtonLamps(id, localWorldView)
		}
	}
}

func drainAndSend(assignmentCh chan [config.N_FLOORS][config.N_BUTTONS]bool, reqs [config.N_FLOORS][config.N_BUTTONS]bool) {
	select {
	case <-assignmentCh:
	default:
	}
	assignmentCh <- reqs
}
