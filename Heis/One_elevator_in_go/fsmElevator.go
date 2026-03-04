package main
type ElevatorIO interface {
	SetMotorDirection(dir Dirn)
	SetDoorOpenLamp(on bool)
	SetFloorIndicator(floor int)
	SetButtonLamp(floor int, btn Button, on bool)
}

type ButtonEvent struct {
	Floor  int
	Button Button
}

func setAllLights(io ElevatorIO, e Elevator) {
	for f := 0; f < N_FLOORS; f++ {
		for b := 0; b < N_BUTTONS; b++ {
			io.SetButtonLamp(f, Button(b), e.requests[f][b])
		}
	}
}

/*
func FindFloor(io ElevatorIO, floorCH chan int) int {
	io.SetMotorDirection(D_Up)


	switchTimer := time.NewTimer(3 * time.Second)
	defer switchTimer.Stop()

	for {
		select {
		case <-switchTimer.C:
			io.SetMotorDirection(D_Down)
		case floor := <-floorCh:
			io.SetMotorDirection(D_Stop)
			return floor

		}
	}
}
*/

func RunFSM(io ElevatorIO, buttonCh <-chan ButtonEvent, floorCh <-chan int, obstrCh <-chan bool) {
	e := elevatorInit()
	doorTimer := NewTimer()

	if e.floor == -1 {
		io.SetMotorDirection(D_Down)
		e.dirn = D_Down
		e.state = ES_Moving
	}

	for {
		select {

		case btn := <-buttonCh:
			onButton(io, doorTimer, &e, btn)
		case newFloor := <-floorCh:
			onFloor(io, doorTimer, &e, newFloor)
		case <-doorTimer.C():
			onDoorTimeout(io, doorTimer, &e)
		case obstr := <-obstrCh:
			if obstr && e.state == ES_DoorOpen {
				doorTimer.Start(e.config.DoorOpenDurations)
			}
		}
	}
}
func onButton(io ElevatorIO, doorTimer *Timer, e *Elevator, ev ButtonEvent) {
	if ev.Floor < 0 || ev.Floor >= N_FLOORS {
		return
	}

	switch e.state {

	case ES_DoorOpen:
		if ClearNow(*e, ev.Floor, ev.Button) {
			doorTimer.Start(e.config.DoorOpenDurations)
		} else {
			e.requests[ev.Floor][ev.Button] = true
		}

	case ES_Moving:
		e.requests[ev.Floor][ev.Button] = true

	case ES_Idle:
		e.requests[ev.Floor][ev.Button] = true

		pair := ChooseDirection(*e)
		e.dirn, e.state = pair.dirn, pair.state

		switch e.state {
		case ES_DoorOpen:
			io.SetDoorOpenLamp(true)
			doorTimer.Start(e.config.DoorOpenDurations)
			ClearAtFloor(e, e.floor)

		case ES_Moving:
			io.SetMotorDirection(e.dirn)

		case ES_Idle:
			//no request anywere, so it stays idle
		}
	}
	setAllLights(io, *e)
}
/* DENNE LAGDE VI PÅ NYTT PÅ TIRSDAG
func onFloor(io ElevatorIO, doorTimer *Timer, e *Elevator, newFloor int) {
	if newFloor < 0 || newFloor >= N_FLOORS {
		return
	}
	e.floor = newFloor
	io.SetFloorIndicator(newFloor)

	if e.state != ES_Moving {
		return
	}
	if ShouldStop(*e) {
		io.SetMotorDirection(D_Stop)
		io.SetDoorOpenLamp(true)
		ClearAtFloor(e, newFloor)
		doorTimer.Start(e.config.DoorOpenDurations)
		setAllLights(io, *e)
		e.state = ES_DoorOpen

	}

}
	*/
func onFloor(io ElevatorIO, doorTimer *Timer, e *Elevator, newFloor int) {
	if newFloor < 0 || newFloor >= N_FLOORS {
		return
	}
	if e.floor == -1 {
		e.floor = newFloor
		io.SetFloorIndicator(newFloor)

		io.SetMotorDirection(D_Stop)
		e.dirn = D_Stop
		e.state = ES_Idle
		io.SetDoorOpenLamp(false)
		

		//setAllLights(io, *e) kanskje ha denne linjenogså for å forsikre seg 
		return
	}
	e.floor = newFloor
	io.SetFloorIndicator(newFloor)

	if e.state != ES_Moving {
		return
	}
	if ShouldStop(*e) {
		io.SetMotorDirection(D_Stop)
		io.SetDoorOpenLamp(true)
		ClearAtFloor(e, newFloor)
		doorTimer.Start(e.config.DoorOpenDurations)
		setAllLights(io, *e)
		e.state = ES_DoorOpen

	}

}


func onDoorTimeout(io ElevatorIO, doorTimer *Timer, e *Elevator) {
	if e.state != ES_DoorOpen {
		return
	}

	pair := ChooseDirection(*e)
	e.dirn = pair.dirn
	e.state = pair.state

	switch e.state {
	case ES_DoorOpen:
		io.SetDoorOpenLamp(true)
		doorTimer.Start(e.config.DoorOpenDurations)
		ClearAtFloor(e, e.floor)
		setAllLights(io, *e)

	case ES_Moving:
		io.SetDoorOpenLamp(false)
		io.SetMotorDirection(e.dirn)

	case ES_Idle:
		io.SetDoorOpenLamp(false)
		io.SetMotorDirection(D_Stop)
	}

}
