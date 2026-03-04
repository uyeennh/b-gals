package main

import (
	"Driver-go/elevio"
)

//import elevio "pathtodrive " where is the driver on the pc?

func main() {
	elevio.Init("localhost:15657", N_FLOORS)

	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors := make(chan int)
	drvObstr := make(chan bool)

	buttonCh := make(chan ButtonEvent)
	floorCh := make(chan int)
	obstrCh := make(chan bool)


	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)

	go func() {
		for btn := range drvButtons {
			buttonCh <- ButtonEvent{
				Floor: btn.Floor,
				Button: Button(btn.Button),
			}
		}
	}()

	go func(){
		for floor := range drvFloors {
			floorCh <- floor
		}
	}()

	go func() {
		for obstr := range drvObstr {
			obstrCh <- obstr
		}
	}()
	io := NewIO(elevioDriver{})
	RunFSM(io, buttonCh, floorCh, obstrCh)
}

type elevioDriver struct{}

func (e elevioDriver) SetMotorDirection(dir int) {
	elevio.SetMotorDirection(elevio.MotorDirection(dir))
}

func (e elevioDriver) SetDoorOpenLamp(on bool) {
	elevio.SetDoorOpenLamp(on)
}

func (e elevioDriver) SetFloorIndicator(floor int) {
	elevio.SetFloorIndicator(floor)
}

func (e elevioDriver) SetButtonLamp(button int, floor int, on bool) {
	elevio.SetButtonLamp(elevio.ButtonType(button), floor, on)
}
