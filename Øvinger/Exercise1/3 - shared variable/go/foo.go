// Use `go run foo.go` to run your program

package main

import (
	"fmt"
	"runtime"
	"sync"
)

var i = 0

func incrementing(inc chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := 0; j < 1000000; j++ {
		inc <- true
	}

}

func decrementing(dec chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := 0; j < 1000000; j++ {
		dec <- true
	}

}

// switch velger hva som er sant, select velger den som er klar
func server(inc chan bool, dec chan bool) {
	for {
		select {
		case <-inc:
			i++
		case <-dec:
			i--
		default:
			return
		}
	}
}

func main() {
	// What does GOMAXPROCS do? What happens if you set it to 1?
	// It uses for concurrency, meaning that we can decide how many threads we can run in parallel.
	// If we set it to 1, then Go will only run one thread at a time, but we can switch between the threads.

	wg := new(sync.WaitGroup)
	wg.Add(2)
	runtime.GOMAXPROCS(3)

	inc := make(chan bool, 1000000)
	dec := make(chan bool, 1000000)

	go server(inc, dec)
	go incrementing(inc, wg)
	go decrementing(dec, wg)

	wg.Wait()

	// TODO: Spawn both functions as goroutines

	// We have no direct way to wait for the completion of a goroutine (without additional synchronization of some sort)
	// We will do it properly with channels soon. For now: Sleep.
	/// time.Sleep(500 * time.Millisecond)

	fmt.Println("The magic number is:", i)
}

/*

func execute() {
	wg := new(sync.WaitGroup)
	wg.Add(3)
	runtime.GOMAXPROCS(3)

	inc := make(chan bool, 1000000)
	dec := make(chan bool, 1000000)

	go server(inc, dec, wg)
	go incrementing(inc, wg)
	go decrementing(dec, wg)

	wg.Wait()
}

func main() {
	execute()
	fmt.Println("The magic number is:", i)
}
*/
