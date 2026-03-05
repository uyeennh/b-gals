package distributor

// ALT ER KOPIERT FRA CLAUDE, IKKE SETT GJENNOM I DET HELE TATT

import (
	"fmt"
	"time"

	"Driver-go/elevio"
	"Network-go/network/bcast"
	"Network-go/network/peers"
	"heis/WorldView"
	"heis/order"
)

// DistributorMessage is what gets broadcast over UDP to all other elevators.
// It contains the sender's ID and their full WorldView.
type DistributorMessage struct {
	Id        string
	WorldView WorldView.WorldView
}

// Distributor is the core synchronization module.
// It owns the WorldView and keeps it consistent across all elevators on the network.
// It runs as a goroutine and communicates via channels.
//
// Inputs:
//
//	id           - unique ID of this elevator (e.g. local IP)
//	numFloors    - number of floors in the building
//	stateCh      - receives physical state updates from the FSM
//	finReqCh     - receives finished requests from the FSM when a floor is served
//	worldViewCh  - sends updated WorldView to the FSM
//	peerUpdateCh - receives peer join/leave events from the network
func Distributor(
	id string,
	numFloors int,
	stateCh <-chan WorldView.ElevatorState,
	finReqCh <-chan elevio.ButtonEvent,
	worldViewCh chan<- WorldView.WorldView,
	peerUpdateCh <-chan peers.PeerUpdate,
) {
	// Initialize our local WorldView
	localWV := WorldView.InitWorldView(id, numFloors)

	// peersAlive is updated by peers.Receiver and used by CyclicCounter
	// to know which elevators must acknowledge before a status advances
	var peersAlive []string

	// Network channels for broadcasting our WorldView to others
	// and receiving their WorldViews
	distMsgTx := make(chan DistributorMessage)
	distMsgRx := make(chan DistributorMessage)
	go bcast.Transmitter(19002, distMsgTx)
	go bcast.Receiver(19002, distMsgRx)

	// Button polling — the Distributor owns buttons, not the FSM.
	// When a button is pressed, the Distributor sets it to OS_Unconf
	// and begins the consensus process.
	drvButtons := make(chan elevio.ButtonEvent)
	go elevio.PollButtons(drvButtons)

	// Broadcast our WorldView every 50ms so other elevators can merge it.
	// This also acts as a heartbeat — if our state changes, others will
	// see it within 50ms even if a UDP packet was lost.
	broadcastTicker := time.NewTicker(50 * time.Millisecond)
	defer broadcastTicker.Stop()

	// Update button lamps every 200ms based on confirmed orders
	lampTicker := time.NewTicker(200 * time.Millisecond)
	defer lampTicker.Stop()

	for {
		select {

		// ── Broadcast our WorldView to all other elevators ───────────────
		case <-broadcastTicker.C:
			distMsgTx <- DistributorMessage{
				Id:        id,
				WorldView: copyWorldView(localWV),
			}

		// ── Receive and merge a WorldView from another elevator ──────────
		case msg := <-distMsgRx:
			// Ignore our own broadcast — we receive what we send on UDP
			if msg.Id == id {
				continue
			}

			// Merge hall orders using CyclicCounter
			localWV.HallOrders = updateHallOrders(
				id,
				localWV.HallOrders,
				msg.WorldView.HallOrders,
				peersAlive,
			)

			// Merge elevator states (physical state + cab orders)
			localWV.States = updateStates(
				id,
				localWV.States,
				msg.Id,
				msg.WorldView.States,
				peersAlive,
			)

			// Send updated WorldView to FSM so it can re-evaluate assignments
			worldViewCh <- copyWorldView(localWV)

		// ── A button was pressed locally ─────────────────────────────────
		case btn := <-drvButtons:
			localWV = handleButtonPress(id, localWV, btn, peersAlive)
			worldViewCh <- copyWorldView(localWV)

		// ── FSM finished serving a request ───────────────────────────────
		case btn := <-finReqCh:
			localWV = handleFinishedOrder(id, localWV, btn, peersAlive)
			worldViewCh <- copyWorldView(localWV)

		// ── FSM published a new physical state ───────────────────────────
		case state := <-stateCh:
			s := localWV.States[id]
			s.Floor = state.Floor
			s.Direction = state.Direction
			s.Behaviour = state.Behaviour
			localWV.States[id] = s

		// ── A peer joined or left the network ────────────────────────────
		case pu := <-peerUpdateCh:
			peersAlive = pu.Peers
			fmt.Println("Peers alive:", peersAlive)

			// When a peer is lost, their unfinished hall orders stay OS_Conf
			// and will be reassigned by the cost function to remaining elevators.
			// We send an updated WorldView to FSM to trigger reassignment.
			worldViewCh <- copyWorldView(localWV)

		// ── Update button lamps ───────────────────────────────────────────
		case <-lampTicker.C:
			updateButtonLamps(localWV, numFloors)
		}
	}
}

// handleButtonPress sets an order to OS_Unconf when a button is pressed.
// If there is only one elevator alive, it immediately confirms the order
// since no other nodes need to acknowledge it.
func handleButtonPress(id string, wv WorldView.WorldView, btn elevio.ButtonEvent, peersAlive []string) WorldView.WorldView {
	switch btn.Button {
	case elevio.BT_Cab:
		// Cab orders belong to this specific elevator
		state := wv.States[id]
		if state.CabOrders[btn.Floor].Status <= order.OS_None {
			state.CabOrders[btn.Floor].Status = order.OS_Unconf
			if len(peersAlive) <= 1 {
				state.CabOrders[btn.Floor].Status = order.OS_Conf
			}
			wv.States[id] = state
		}
	default:
		// Hall orders are shared across all elevators
		b := int(btn.Button)
		if wv.HallOrders[btn.Floor][b].Status <= order.OS_None {
			wv.HallOrders[btn.Floor][b].Status = order.OS_Unconf
			if len(peersAlive) <= 1 {
				wv.HallOrders[btn.Floor][b].Status = order.OS_Conf
			}
		}
	}
	return wv
}

// handleFinishedOrder sets an order to OS_Fin when the FSM has served it.
// If there is only one elevator alive, it immediately clears the order
// since no other nodes need to acknowledge the completion.
func handleFinishedOrder(id string, wv WorldView.WorldView, btn elevio.ButtonEvent, peersAlive []string) WorldView.WorldView {
	switch btn.Button {
	case elevio.BT_Cab:
		state := wv.States[id]
		state.CabOrders[btn.Floor].Status = order.OS_Fin
		if len(peersAlive) <= 1 {
			state.CabOrders[btn.Floor].Status = order.OS_None
		}
		wv.States[id] = state
	default:
		b := int(btn.Button)
		wv.HallOrders[btn.Floor][b].Status = order.OS_Fin
		if len(peersAlive) <= 1 {
			wv.HallOrders[btn.Floor][b].Status = order.OS_None
		}
	}
	return wv
}

// updateHallOrders merges the local and received hall orders
// for every floor and button using CyclicCounter.
func updateHallOrders(id string, local [][2]order.Order, received [][2]order.Order, peersAlive []string) [][2]order.Order {
	for f := 0; f < len(local); f++ {
		for b := 0; b < 2; b++ {
			local[f][b] = order.CyclicCounter(id, local[f][b], received[f][b], peersAlive)
		}
	}
	return local
}

// updateStates merges elevator states from a received WorldView into the local one.
// Physical state (floor, direction, behaviour) is taken directly from the sender.
// Cab orders are merged using CyclicCounter since they need consensus too.
func updateStates(
	localId string,
	localStates map[string]WorldView.ElevatorState,
	recvId string,
	recvStates map[string]WorldView.ElevatorState,
	peersAlive []string,
) map[string]WorldView.ElevatorState {

	// Update the physical state of the elevator that sent this message
	if recvState, ok := recvStates[recvId]; ok {
		if existing, exists := localStates[recvId]; exists {
			existing.Floor = recvState.Floor
			existing.Direction = recvState.Direction
			existing.Behaviour = recvState.Behaviour
			localStates[recvId] = existing
		} else {
			// First time we hear from this elevator — add it to our WorldView
			localStates[recvId] = recvState
		}
	}

	// Synchronize cab orders for all elevators we know about
	// Cab orders need consensus too — if elevator A crashes,
	// we need to know its cab orders so they can be reassigned
	for recvStateId, recvState := range recvStates {
		local, exists := localStates[recvStateId]
		if !exists {
			continue
		}
		for f := 0; f < len(recvState.CabOrders); f++ {
			local.CabOrders[f] = order.CyclicCounter(
				localId,
				local.CabOrders[f],
				recvState.CabOrders[f],
				peersAlive,
			)
		}
		localStates[recvStateId] = local
	}

	return localStates
}

// updateButtonLamps turns hall button lamps on when an order is OS_Conf or higher,
// and off when it returns to OS_None.
func updateButtonLamps(wv WorldView.WorldView, numFloors int) {
	for f := 0; f < numFloors; f++ {
		// Hall up lamp
		elevio.SetButtonLamp(elevio.BT_HallUp, f,
			wv.HallOrders[f][0].Status >= order.OS_Conf)
		// Hall down lamp
		elevio.SetButtonLamp(elevio.BT_HallDown, f,
			wv.HallOrders[f][1].Status >= order.OS_Conf)
	}
}

// copyWorldView creates a deep copy of the WorldView to avoid race conditions
// when sending it over a channel while the distributor continues modifying it.
func copyWorldView(wv WorldView.WorldView) WorldView.WorldView {
	newWV := WorldView.WorldView{
		HallOrders: make([][2]order.Order, len(wv.HallOrders)),
		States:     make(map[string]WorldView.ElevatorState),
	}
	copy(newWV.HallOrders, wv.HallOrders)
	for k, v := range wv.States {
		newWV.States[k] = v
	}
	return newWV
}
