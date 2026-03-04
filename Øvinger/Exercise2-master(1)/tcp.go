package main

import (
	"bufio"
	"fmt"
	"net"
)

const (
	SERVER_IP = "10.100.23.11"
	PORT      = "33546"
	MY_IP     = "10.100.23.14"
)

func connectTCP(serverIP string, port string) error {

	conn, err := net.Dial("tcp", serverIP+":"+port)
	if err != nil {
		fmt.Println("Failed to connect", err)
		return err

	}

	fmt.Println("Connected to server at", serverIP+":"+port)

	reader := bufio.NewReader(conn)

	welcome, err := reader.ReadString('\x00')
	if err != nil {
		fmt.Println("Failed to read", err)
		return err
	}

	fmt.Println("Server says:", welcome[:len(welcome)-1])

	message := fmt.Sprintf("Connect to: %s:%s\x00", MY_IP, port)

	_, err = conn.Write([]byte(message))
	// if we get an error message, then the whole message was not handed.
	if err != nil {
		fmt.Println("Failed to write", err)
		return err
	}

	return nil
}

// er bruk av ready en dårlig måte å løse raceconditio problematikken eller burde vi helelr bruke en sleep
func listencallback(port string, ready chan<- struct{}) {
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Failed to listen", err)
		return
	}
	defer listen.Close()
	fmt.Println("Listening for server callback on port", port)

	close(ready)
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Failed to listen", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\x00') // read until null terminator
		if err != nil {
			fmt.Println("Connection closed:", err)
			return
		}
		fmt.Printf("Received: %s\n", msg[:len(msg)-1])

	}
}

/*
func main() {
	ready := make(chan struct{})
	go listencallback(PORT, ready)

	<-ready

	err := connectTCP(SERVER_IP, PORT)

	if err != nil {
		fmt.Println("Error:", err)
	}
	select {}
}
*/
