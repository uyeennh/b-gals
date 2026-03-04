package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	go reciever()

	go sender()

	select {}

}

func reciever() {
	const listen_port = 30000
	const server_type = "udp"
	// Turns an IP and port (as a string) into a UDP address object that Go can use.
	addr, err := net.ResolveUDPAddr(server_type, fmt.Sprintf(":%d", listen_port))
	if err != nil {
		fmt.Println("NOT FOUND address", err)
		return
	}

	// Create a socket
	conn, err := net.ListenUDP(server_type, addr)
	if err != nil {
		fmt.Println("Failed to listen", err)
		return
	}
	defer conn.Close()

	fmt.Printf("Receiver listening on port %d...\n", listen_port)

	buffer := make([]byte, 1024)

	// Receive messages in a loop
	for {
		n, remote_addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading data:", err)
			continue
		}
		// the buffer just contains a bunch of bytes, so you may have to explicitly convert it to a string
		message := string(buffer[:n])
		fmt.Printf("Received message from %s: %s\n", remote_addr, message)
	}
}

// sender sends a message to a given IP and port over UDP
func sender() {
	time.Sleep(1 * time.Second)

	serverIP := "10.100.23.11"
	serverPort := 30000

	// Turns an IP and port (as a string) into a UDP address object that Go can use.
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", serverIP, serverPort))
	if err != nil {
		fmt.Println("NOT FOUND address", err)
		return
	}

	// Create a socket
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Failed to listen", err)
		return
	}
	defer conn.Close()

	for i := 0; ; i++ {
		message := fmt.Sprintf(" The message is #%d", i)
		_, err = conn.Write([]byte(message))

		if err != nil {
			fmt.Errorf("failed to send message: %w", err)
			continue

		}

		fmt.Printf("Message sent: %s\n", message)
		time.Sleep(2 * time.Second)
	}
}
