

package costfunction


type HRAElevState struct {
	Behavior		string  'json: "behaviour"'
	Floor 			int 	'json: "floor"'
	Direction 		string 	'json: "direction"'
	CabRequests		[]bool	'json: "cabRequests"'
}


type HRAInput struct {
    HallRequests    [][2]bool                   `json:"hallRequests"`
    States          map[string]HRAElevState     `json:"states"`
}

func Compute(
	myID string, 
	hallRequests [][2]bool, 
	states map[string]HRAElevState,
)([][2]bool, bool){
	hraExecutable := ""
    switch runtime.GOOS {
        case "linux":   hraExecutable  = "hall_request_assigner"
        case "windows": hraExecutable  = "hall_request_assigner.exe"
        default:        panic("OS not supported")
    }

	input :=HRAInput{
		HallRequests: 	hallRequests,
		States: 		states,
	}
	jsonBytes, err := json.Marshal(input)
	if err !=nil {
		return nil, false
	}

	ret, err : exec.Command(
		"../hall_request_assigner/"+hraExecutable,
		"-i",
		string(jsonBytes),
	).CombinedOutput()

	if err != nil {
		return nil, false
	}
	
	var output map[string][][2]bool

	err = json.Unmarshal(ret, &output)
	if err != nil {
		return nil, false
	}

	assigned, ok := output[myID]
	if !ok {
		return nil, false 
	}

	return assigned, true
	
}



