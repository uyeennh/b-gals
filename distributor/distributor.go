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
	assignInterval     = 200 * time.Millisecond
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
 
	// assignTicker is the ONLY place assignments are computed.
	// By decoupling assignment from worldview updates, all elevators
	// compute their assignments at a scheduled moment — after multiple
	// broadcast cycles have happened and the worldview has converged.
	// This means all elevators see the same picture and the cost function
	// gives consistent results — only one elevator moves per order.
	assignTicker := time.NewTicker(assignInterval)
	defer assignTicker.Stop()
 
	for {
		select {
 
		case <-broadcastTicker.C:
			// Broadcast our current worldview to all other elevators
			msg := networkMessage{
				SenderID:  id,
				WorldView: copyWorldView(localWorldView),
			}
			select {
			case <-messageTx:
			default:
			}
			messageTx <- msg
 
		case message := <-messageRx:
			if message.SenderID == id {
				continue
			}
			// Only update the worldview — do NOT compute assignments here.
			// The assignTicker handles that separately.
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
 
		case buttonEvent := <-driverButtonCh:
			// Only update the worldview — do NOT compute assignments here.
			localWorldView = handleButtonPress(id, localWorldView, buttonEvent, peersAlive)
 
		case completedOrder := <-finReqCh:
			// Only update the worldview — do NOT compute assignments here.
			localWorldView = handleFinishedOrder(id, localWorldView, completedOrder, peersAlive)
			// Self-merge so our own finished barrier advances immediately
			localWorldView.HallOrders = mergeHallOrders(
				id,
				localWorldView.HallOrders,
				copyHallOrders(localWorldView.HallOrders),
				peersAlive,
			)
 
		case state := <-stateCh:
			// Update our own physical state in the worldview
			elevatorState := localWorldView.States[id]
			elevatorState.Floor = state.Floor
			elevatorState.Direction = state.Direction
			elevatorState.Behaviour = state.Behaviour
			localWorldView.States[id] = elevatorState
			// Immediately broadcast our new state (moving/idle/doorOpen)
			// so other elevators see it before the next assignTicker fires
			select {
			case <-messageTx:
			default:
			}
			messageTx <- networkMessage{
				SenderID:  id,
				WorldView: copyWorldView(localWorldView),
			}
 
		case peerUpdate := <-peerUpdateCh:
			// Only update the peer list — do NOT compute assignments here.
			peersAlive = peerUpdate.Peers
			if !order.ContainsID(peersAlive, id) {
				peersAlive = append(peersAlive, id)
			}
			fmt.Println("Peers alive:", peersAlive)
 
		case <-assignTicker.C:
			// This is the ONLY place we compute and send assignments.
			// By this point, multiple broadcast cycles have happened since
			// any worldview change, so all elevators have consistent state.
			computeAndSendAssignment(id, localWorldView, peersAlive, assignmentCh)
 
		case <-lampTicker.C:
			updateButtonLamps(id, localWorldView)
		}
	}
}