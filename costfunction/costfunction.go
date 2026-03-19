package costfunction

//should you use this somewhere? cmd := exec.Command(config.HRAPath)

import (
	"encoding/json"
	"heis/config"
	"os/exec"
	"runtime"
	"fmt"
)

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

func Compute(
	myID string,
	hallRequests [][2]bool,
	states map[string]HRAElevState,
) ([][2]bool, bool) {

	input := HRAInput{
		HallRequests: hallRequests,
		States:       states,
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return nil, false
	}
	hasHallReqs := false
	for _, pair := range hallRequests {
    	if pair[0] || pair[1] {
        	hasHallReqs = true
        	break
    	}	
	}

	if hasHallReqs {
    	fmt.Printf("[%s] floors:", myID)
    	for id, s := range states {
        	fmt.Printf(" %s(f%d,%s,%s)", id, s.Floor, s.Direction, s.Behavior)
    	}
    	fmt.Printf(" halls:")
    	for f, pair := range hallRequests {
        	if pair[0] || pair[1] {
            	fmt.Printf(" f%d(up=%v,dn=%v)", f, pair[0], pair[1])
        	}
    	}
    	fmt.Println()
	}

	rawOutput, err := exec.Command(executableName(), "-i", string(jsonBytes)).CombinedOutput()
	if err != nil {
		return nil, false
	}

	var result map[string][][2]bool
	if err := json.Unmarshal(rawOutput, &result); err != nil {
		return nil, false
	}

	assigned, ok := result[myID]
	if !ok {
		return nil, false
	}

	return assigned, true

}
func executableName() string {
	switch runtime.GOOS {
	case "linux":
		return config.HRAPath + "hall_request_assigner"
	case "windows":
		return config.HRAPath + "hall_request_assigner.exe"
	default:
		panic("unsupported OS: " + runtime.GOOS)
	}
}

/*
hraExecutable := ""
    switch runtime.GOOS {
        case "linux":   hraExecutable  = "hall_request_assigner"
        case "windows": hraExecutable  = "hall_request_assigner.exe"
        default:        panic("OS not supported")
    }

ret, err := exec.Command(
		"../hall_request_assigner/"+hraExecutable,
		"-i",
		string(jsonBytes),
	).CombinedOutput()

	var output map[string][][2]bool
	err = json.Unmarshal(ret, &output) */
