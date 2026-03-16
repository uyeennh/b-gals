package fsm

import (
	"heis/config"
	"heis/driver/elevio"
)

func RunFSM(
	io ElevatorIO,
	assignmentCh <-chan [config.N_FLOORS][config.N_BUTTONS]bool,
	finReqCh chan<- elevio.ButtonEvent,
	stateCh chan<- ElevatorStateMsg,
	floorCh <-chan int,
	obstrCh <-chan bool,
) {
	e := elevatorInit()
	doorTimer := NewTimer()
	motorTimer := NewTimer()

	obstrActive := false
	obstrCounter := 0
	obstrStoredFloor := -1

	io.SetMotorDirection(D_Down)
	e.dirn = D_Down
	e.state = ES_Moving
	motorTimer.Start(config.MotorLossTimeout)

	for {
		select {

		case assignments := <-assignmentCh:
			e.requests = assignments
			switch e.state {
			case ES_Idle:
				pair := ChooseDirection(e)
				e.dirn, e.state = pair.dirn, pair.state

				switch e.state {
				case ES_DoorOpen:
					io.SetDoorOpenLamp(true)
					doorTimer.Start(e.elevConfig.DoorOpenDuration)
					ClearAtFloor(&e, e.floor, finReqCh)
					setAllLights(io, e)
				case ES_Moving:
					io.SetMotorDirection(e.dirn)
					motorTimer.Start(config.MotorLossTimeout)
				case ES_Idle:
					// nothing to do
				}
			case ES_Moving:
				pair := ChooseDirection(e)
				if pair.state == ES_Idle {
					io.SetMotorDirection(D_Stop)
					motorTimer.Stop()
					e.dirn = D_Stop
					e.state = ES_Idle
				}

			case ES_DoorOpen:
			}

			stateCh <- ElevatorStateMsg{Floor: e.floor, Dirn: e.dirn, State: e.state}

		case newFloor := <-floorCh:
			motorTimer.Stop()

			if e.floor == -1 {
				e.floor = newFloor
				io.SetMotorDirection(D_Stop)
				io.SetFloorIndicator(newFloor)
				io.SetDoorOpenLamp(false)
				e.dirn = D_Stop
				e.state = ES_Idle
				// Check if there is already a request at this floor
				if ShouldStop(e) {
					io.SetDoorOpenLamp(true)
					ClearAtFloor(&e, newFloor, finReqCh)
					doorTimer.Start(e.elevConfig.DoorOpenDuration)
					e.state = ES_DoorOpen
				} else {
					io.SetDoorOpenLamp(false)
				}

				stateCh <- ElevatorStateMsg{Floor: e.floor, Dirn: e.dirn, State: e.state}
				continue
			}

			e.floor = newFloor
			io.SetFloorIndicator(newFloor)

			if e.state != ES_Moving {
				stateCh <- ElevatorStateMsg{Floor: e.floor, Dirn: e.dirn, State: e.state}
				continue
			}

			if ShouldStop(e) {
				io.SetMotorDirection(D_Stop)
				io.SetDoorOpenLamp(true)
				ClearAtFloor(&e, newFloor, finReqCh)
				doorTimer.Start(e.elevConfig.DoorOpenDuration)
				setAllLights(io, e)
				e.state = ES_DoorOpen
			} else {
				motorTimer.Start(config.MotorLossTimeout)
			}

			stateCh <- ElevatorStateMsg{Floor: e.floor, Dirn: e.dirn, State: e.state}

		case <-doorTimer.C():
			if e.state != ES_DoorOpen {
				continue
			}

			if obstrActive {
				doorTimer.Start(e.elevConfig.DoorOpenDuration)

				obstrCounter++
				if obstrCounter >= config.ObstrTripsBeforeFloorInvalid {
					obstrStoredFloor = e.floor
					e.floor = -1
					stateCh <- ElevatorStateMsg{Floor: -1, Dirn: e.dirn, State: e.state}
				}
				continue
			}

			pair := ChooseDirection(e)
			e.dirn, e.state = pair.dirn, pair.state

			switch e.state {
			case ES_DoorOpen:
				io.SetDoorOpenLamp(true)
				doorTimer.Start(e.elevConfig.DoorOpenDuration)
				ClearAtFloor(&e, e.floor, finReqCh)
				setAllLights(io, e)

			case ES_Moving:
				io.SetDoorOpenLamp(false)
				io.SetMotorDirection(e.dirn)
				motorTimer.Start(config.MotorLossTimeout)

			case ES_Idle:
				io.SetDoorOpenLamp(false)
				io.SetMotorDirection(D_Stop)
			}

			stateCh <- ElevatorStateMsg{Floor: e.floor, Dirn: e.dirn, State: e.state}

		case <-motorTimer.C():
			e.floor = -1
			stateCh <- ElevatorStateMsg{Floor: -1, Dirn: e.dirn, State: e.state}

		case obstr := <-obstrCh:
			obstrActive = obstr

			if obstr {
				if e.state == ES_DoorOpen {
					doorTimer.Start(e.elevConfig.DoorOpenDuration)
				}
			} else {
				if obstrCounter >= config.ObstrTripsBeforeFloorInvalid {
					e.floor = obstrStoredFloor
					stateCh <- ElevatorStateMsg{Floor: e.floor, Dirn: e.dirn, State: e.state}
				}
				obstrCounter = 0
				obstrStoredFloor = -1
			}
		}
	}
}

func setAllLights(io ElevatorIO, e Elevator) {
	for f := 0; f < config.N_FLOORS; f++ {
		io.SetButtonLamp(f, B_Cab, e.requests[f][B_Cab])
	}
}
