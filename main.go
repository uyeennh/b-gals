package main


//forslag til main.go fra claude. må gås gjennom og optimaliserers mer
// gamle main ligger under øvinger en plass i tilfelle det var noe vi trenger der



// main.go — entry point for the heis (elevator) node.
//
// Responsibilities:
//   1. Parse the elevator ID from the command line (or fall back to local IP).
//   2. Initialise the hardware driver.
//   3. Spin up the peer-discovery heartbeat (UDP broadcast).
//   4. Wire all channels between the FSM and the Distributor.
//   5. Start the Distributor and FSM goroutines.
//
// Channel flow:
//
//   elevio pollers ──► floorCh / obstrCh
//                          │
//                          ▼
//                        RunFSM ──► stateMsgCh (fsm.ElevatorStateMsg)
//                          │               │
//                          │        stateAdapter goroutine
//                          │               │
//                          │               ▼
//                          │        stateCh (WorldView.ElevatorState)
//                          │               │
//                          ▼               ▼
//                        finReqCh ──► Distributor ──► assignmentCh ──► RunFSM
//                                         ▲
//                                   peerUpdateCh (peers.Receiver)

import (
	"fmt"
	"os"
	"log"

	"heis/WorldView"
	"heis/config"
	"heis/distributor"
	"heis/driver/elevio"
	"heis/elevtype"
	"heis/network/peers"
	fsm "heis/fsm"
)

func main() {
	// ── 1. Elevator ID ────────────────────────────────────────────────────
	// Pass an explicit ID on the command line, e.g.  ./heis elev1
	// If none is given, derive one from the machine's local IP so that three
	// elevators running on three different machines get unique IDs without
	// any manual configuration.
	//id := resolveID(os.Args)
	if len(os.Args) < 2 {
		log.Fatal("Usage: ./heis <config-file>  e.g.  ./heis config/elevator1.con")
	}
	conf := config.ReadConfigFile(os.Args[1])
	id := conf.Id
	fmt.Printf("[main] Starting elevator  id=%q  port=%d\n", id, conf.PortIo)


	// ── 2. Hardware driver ────────────────────────────────────────────────

	elevio.Init(fmt.Sprintf("localhost:%d", conf.PortIo), config.N_FLOORS)
	// rest of main stays exactly the same


	fmt.Printf("[main] Starting elevator node  id=%q\n", id)


	// ── 3. Peer-discovery (heartbeat) ─────────────────────────────────────
	// peers.Transmitter broadcasts our ID on UDP every 15 ms.
	// peers.Receiver listens and emits PeerUpdate events when nodes join or leave.
	peerUpdateCh  := make(chan peers.PeerUpdate)
	transmitEnable := make(chan bool, 1)

	go peers.Transmitter(config.PortPeers, id, transmitEnable)
	go peers.Receiver(config.PortPeers, peerUpdateCh)

	// Enable heartbeat transmission immediately.
	transmitEnable <- true

	// ── 4. Channel declarations ───────────────────────────────────────────

	// FSM  →  Distributor: physical state updates after every event.
	// RunFSM sends fsm.ElevatorStateMsg; Distributor wants WorldView.ElevatorState.
	// A lightweight adapter goroutine (see below) bridges the two types.
	stateMsgCh := make(chan fsm.ElevatorStateMsg, 4)
	stateCh    := make(chan WorldView.ElevatorState, 4)

	// Distributor  →  FSM: assigned request matrix after every re-computation.
	// Buffered so that a slow FSM does not block the Distributor's select loop.
	assignmentCh := make(chan [config.N_FLOORS][config.N_BUTTONS]bool, 1)

	// FSM  →  Distributor: completed orders (floor + button) so the WorldView
	// can mark them as done and clear the corresponding lamps.
	finReqCh := make(chan elevio.ButtonEvent, 8)

	// Hardware pollers  →  FSM
	floorCh := make(chan int)
	obstrCh := make(chan bool)

	// ── 5. Hardware pollers ───────────────────────────────────────────────
	go elevio.PollFloorSensor(floorCh)
	go elevio.PollObstructionSwitch(obstrCh)
	// Button polling is handled inside Distributor (it needs raw button events
	// to register new hall and cab orders in the WorldView before assigning them).

	// ── 6. State type adapter ─────────────────────────────────────────────
	// fsm.ElevatorStateMsg  →  WorldView.ElevatorState
	//
	// The FSM uses its own internal ElevatorState enum (ES_Idle / ES_Moving /
	// ES_DoorOpen) while the WorldView uses elevtype.Behaviour.  This goroutine
	// translates between the two so neither package needs to import the other.
	go func() {
		for msg := range stateMsgCh {
			stateCh <- WorldView.ElevatorState{
				Floor:     msg.Floor,
				Direction: elevtype.Dirn(msg.Dirn),
				Behaviour: fsmStateToBehaviour(msg.State),
				// CabOrders are managed by the Distributor;
				// we do not overwrite them here.
			}
		}
	}()

	// ── 7. IO adapter (elevio ↔ fsm.Driver interface) ─────────────────────
	io := fsm.NewIO(elevioAdapter{})

	// ── 8. Launch goroutines ──────────────────────────────────────────────
	go distributor.Distributor(id, stateCh, finReqCh, assignmentCh, peerUpdateCh)

	// RunFSM blocks — keep it on the main goroutine so the process stays alive.
	fsm.RunFSM(io, assignmentCh, finReqCh, stateMsgCh, floorCh, obstrCh)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// resolveID returns the first command-line argument after the binary name,
// or falls back to a string derived from the local IP address.
/*

func resolveID(args []string) string {
	if len(args) >= 2 && args[1] != "" {
		return args[1]
	}
	return localIPFallback()
}

// localIPFallback dials a UDP connection (never sends anything) to discover
// the machine's outbound IP and uses its last two octets as an ID.
// This gives stable, unique IDs like "1.42" across machines on the same LAN
// without requiring manual setup.
func localIPFallback() string {
	conn, err := net.Dial("udp4", "8.8.8.8:80")
	if err != nil {
		return "unknown"
	}
	defer conn.Close()
	addr := conn.LocalAddr().(*net.UDPAddr)
	ip := addr.IP.To4()
	if ip == nil {
		return "unknown"
	}
	return fmt.Sprintf("%d.%d", ip[2], ip[3])
}
*/

// fsmStateToBehaviour converts the FSM-internal ElevatorState enum to the
// elevtype.Behaviour enum that the WorldView and Distributor understand.
func fsmStateToBehaviour(s fsm.ElevatorState) elevtype.Behaviour {
	switch s {
	case fsm.ES_Moving:
		return elevtype.B_Moving
	case fsm.ES_DoorOpen:
		return elevtype.B_DoorOpen
	default: // ES_Idle and any unknown state
		return elevtype.B_Idle
	}
}

// ── elevioAdapter ─────────────────────────────────────────────────────────────
// Thin wrapper that satisfies the fsm.Driver interface using the elevio package.
// Keeps main.go as the only place that knows about both packages — neither
// fsm nor elevio imports the other.

type elevioAdapter struct{}

func (elevioAdapter) SetMotorDirection(dir int) {
	elevio.SetMotorDirection(elevio.MotorDirection(dir))
}

func (elevioAdapter) SetDoorOpenLamp(on bool) {
	elevio.SetDoorOpenLamp(on)
}

func (elevioAdapter) SetFloorIndicator(floor int) {
	elevio.SetFloorIndicator(floor)
}

func (elevioAdapter) SetButtonLamp(button int, floor int, on bool) {
	elevio.SetButtonLamp(elevio.ButtonType(button), floor, on)
}
