package main

import (
	"fmt"
	"log"
	"os"

	"heis/config"
	"heis/distributor"
	"heis/driver/elevio"
	"heis/elevtype"
	"heis/fsm"
	"heis/network/peers"
	"heis/worldview"
)

const (
	stateChannelBufferSize    = 4
	finishedRequestBufferSize = 8
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ./heis <config-file>  e.g.  ./heis config/elevator1.con")
	}

	conf := config.ReadConfigFile(os.Args[1])
	id := conf.Id
	fmt.Printf("[main] Starting elevator  id=%q  port=%d\n", id, conf.PortIo)

	elevio.Init(fmt.Sprintf("localhost:%d", conf.PortIo), config.N_FLOORS)

	peerUpdateCh := make(chan peers.PeerUpdate)
	transmitEnable := make(chan bool, 1)

	go peers.Transmitter(config.PortPeers, id, transmitEnable)
	go peers.Receiver(config.PortPeers, peerUpdateCh)
	transmitEnable <- true

	stateMsgCh := make(chan fsm.ElevatorStateMsg, stateChannelBufferSize)
	stateCh := make(chan worldview.ElevatorState, stateChannelBufferSize)
	assignmentCh := make(chan [config.N_FLOORS][config.N_BUTTONS]bool, 1)
	finReqCh := make(chan elevio.ButtonEvent, finishedRequestBufferSize)
	floorCh := make(chan int)
	obstrCh := make(chan bool)

	go elevio.PollFloorSensor(floorCh)
	go elevio.PollObstructionSwitch(obstrCh)
	go translateFSMStateToWorldView(stateMsgCh, stateCh)

	io := fsm.NewIO(elevioAdapter{})

	go distributor.Distributor(id, stateCh, finReqCh, assignmentCh, peerUpdateCh)
	fsm.RunFSM(io, assignmentCh, finReqCh, stateMsgCh, floorCh, obstrCh)
}

func translateFSMStateToWorldView(in <-chan fsm.ElevatorStateMsg, out chan<- worldview.ElevatorState) {
	for msg := range in {
		out <- worldview.ElevatorState{
			Floor:     msg.Floor,
			Direction: elevtype.Dirn(msg.Dirn),
			Behaviour: fsmStateToBehaviour(msg.State),
		}
	}
}

func fsmStateToBehaviour(s fsm.ElevatorState) elevtype.Behaviour {
	switch s {
	case fsm.ES_Moving:
		return elevtype.B_Moving
	case fsm.ES_DoorOpen:
		return elevtype.B_DoorOpen
	default:
		return elevtype.B_Idle
	}
}

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