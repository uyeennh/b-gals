package distributor

import (
	"fmt"
	"heis/worldview"
	"heis/config"
	"heis/driver/elevio"
	"heis/network/bcast"
	"heis/network/peers"
	"heis/order"
	"heis/saveIfCrash"
	"time"
)

const (
	broadcastInterval  = 50 * time.Millisecond
	lampUpdateInterval = 200 * time.Millisecond
)



type networkMessage struct {
	SenderID  string
	WorldView  worldview.WorldView
}

func Distributor(
	id string,
	stateCh <-chan worldview.ElevatorState,
	finReqCh <-chan elevio.ButtonEvent,
	assignmentCh chan [config.N_FLOORS][config.N_BUTTONS]bool,
	peerUpdateCh <-chan peers.PeerUpdate,
) {
	localWV := worldview.InitWorldView(id, config.N_FLOORS)
	savedCabOrders := saveIfCrash.LoadCabOrders(id, config.N_FLOORS)
	for f := range savedCabOrders {
		if savedCabOrders[f].Status >= order.OS_Unconst {
			savedCabOrders[f].Status = order.OS_Const
			savedCabOrders[f].Barrier = []string{}
		}
	}
	restoredState := localWV.States[id]
	restoredState.CabOrders = savedCabOrders
	localWV.States[id] = restoredState

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
				WorldView: copyWorldView(localWV),
			}

		case msg := <-msgRx:
			if msg.SenderID == id {
				continue
			}

			localWV.HallOrders = mergeHallOrders(
				id,
				localWV.HallOrders,
				msg.worldview.HallOrders,
				peersAlive,
			)

			localWV.States = mergeStates(
				id,
				localWV.States,
				msg.SenderID,
				msg.worldview.States,
				peersAlive,
			)

			computeAndSendAssignment(id, localWV, peersAlive, assignmentCh)

		case btn := <-drvButtons:
			localWV = handleButtonPress(id, localWV, btn, peersAlive)
			computeAndSendAssignment(id, localWV, peersAlive, assignmentCh)
		case btn := <-finReqCh:
			localWV = handleFinishedOrder(id, localWV, btn, peersAlive)
			// Self-merge so our own barrier advances immediately
			localWV.HallOrders = mergeHallOrders(
				id,
				localWV.HallOrders,
				copyHallOrders(localWV.HallOrders),
				peersAlive,
			)
			computeAndSendAssignment(id, localWV, peersAlive, assignmentCh)

		case state := <-stateCh:
			s := localWV.States[id]
			s.Floor = state.Floor
			s.Direction = state.Direction
			s.Behaviour = state.Behaviour
			localWV.States[id] = s

		case pu := <-peerUpdateCh:
			peersAlive = pu.Peers
			// Always make sure our own ID is in the peer list
			// peers.Receiver never includes ourselves because we never
			// hear our own broadcast — so we add ourselves manually here
			if !order.ContainsID(peersAlive, id) {
				peersAlive = append(peersAlive, id)
			}
			fmt.Println("Peers alive:", peersAlive)
			computeAndSendAssignment(id, localWV, peersAlive, assignmentCh)
		case <-lampTicker.C:
			updateButtonLamps(id, localWV)
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
