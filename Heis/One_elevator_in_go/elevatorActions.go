package main



type Driver interface {
	SetMotorDirection(dir int)                 
	SetDoorOpenLamp(on bool)
	SetFloorIndicator(floor int)
	SetButtonLamp(button int, floor int, on bool) 
}

type IO struct {
	drv Driver
}

func NewIO(drv Driver) *IO {
	return &IO{drv: drv}
}


func (io *IO) SetMotorDirection(dir Dirn) {
	if io == nil || io.drv == nil {
		return
	}
	//kan skrives som io.drv.SetMotorDirection(int(dir))
	io.drv.SetMotorDirection(dirnToDriver(dir))
}

func (io *IO) SetDoorOpenLamp(on bool) {
	if io == nil || io.drv == nil {
		return
	}
	io.drv.SetDoorOpenLamp(on)
}

func (io *IO) SetFloorIndicator(floor int) {
	if io == nil || io.drv == nil {
		return
	}
	if floor < 0 || floor >= N_FLOORS {
		return
	}
	io.drv.SetFloorIndicator(floor)
}

func (io *IO) SetButtonLamp(floor int, btn Button, on bool) {
	if io == nil || io.drv == nil {
		return
	}
	if floor < 0 || floor >= N_FLOORS {
		return
	}
	io.drv.SetButtonLamp(buttonToDriver(btn), floor, on)
}
func elevatorInit() Elevator{
	return Elevator{
		floor: -1,
		dirn: D_Stop,
		requests: [N_FLOORS][N_BUTTONS]bool{},
		state: ES_Idle,
		config: Config{DoorOpenDurations: 3e9},
	}
}



func dirnToDriver(d Dirn) int {
	
	return int(d)
}

func buttonToDriver(b Button) int {
	
	return int(b)
}




