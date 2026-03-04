package main

import (
	"fmt"
	"time"
)

func producer(buffer chan int) {

	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Printf("[producer]: pushing %d\n", i)
		// TODO: push real value to buffer by sending value to channel
		buffer <- i
	}

}

func consumer(buffer chan int) {

	time.Sleep(5 * time.Second)
	for {
		//i := <-buffer //TODO: get real value from buffer
		fmt.Printf("[consumer]: %d\n", <-buffer)
		time.Sleep(50 * time.Millisecond)
	}

}

// channel prevents producert ot push into full channel and consumer to pop from emtpy channel
// channel replace mutex and semaphores

func main() {

	// TODO: make a bounded buffer
	buffer := make(chan int, 5)

	go consumer(buffer)
	go producer(buffer)

	select {}
}
