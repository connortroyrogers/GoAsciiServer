package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	MESSAGELIMIT = 256
)

func main() {
	ip := "127.0.0.1"
	port := 4040
	mode := flag.String("mode", "", "startup mode: server or client")
	flag.Parse()

	choice := strings.ToLower(strings.TrimSpace(*mode))
	if choice == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Enter 1 to set up client, 2 to set up server")
		tempchoice, _ := reader.ReadString('\n')
		rawChoice := strings.TrimSpace(tempchoice)
		switch rawChoice {
		case "1":
			choice = "client"
		case "2":
			choice = "server"
		default:
			choice = rawChoice
		}
	}

	fmt.Printf("Using %s:%d\n", ip, port)

	switch choice {
	case "1", "client":
		gbnClient(ip, port)
	case "2", "server":
		gbnServer(ip, port)
	default:
		fmt.Println("Error incorrect selection. Use -mode=server or -mode=client")
	}
}

// Server
func gbnServer(iface string, port int) {
	addr := net.UDPAddr{
		IP:   net.ParseIP(iface),
		Port: port,
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Printf("Server listening on %s:%d\n", iface, port)
	fmt.Println("Waiting for client to connect...")

	handshakeBuf := make([]byte, MESSAGELIMIT)
	_, clientAddr, err := conn.ReadFromUDP(handshakeBuf)
	if err != nil {
		fmt.Println("handshake error:", err)
		return
	}

	fmt.Printf("Client connected: %s\n", clientAddr.String())
	fmt.Println("Type a message to send to client (type 'exit' to quit)")

	stdin := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("server> ")
		text, err := stdin.ReadString('\n')
		if err != nil {
			fmt.Println("input error:", err)
			return
		}

		message := strings.TrimSpace(text)
		if message == "" {
			continue
		}

		_, err = conn.WriteToUDP([]byte(message), clientAddr)
		if err != nil {
			fmt.Println("write error:", err)
			continue
		}

		if message == "exit" {
			return
		}
	}
}

// Client
func gbnClient(host string, port int) {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("client-hello"))
	if err != nil {
		fmt.Println("handshake send error:", err)
		return
	}

	fmt.Println("Connected to server. Waiting for messages...")
	buf := make([]byte, MESSAGELIMIT)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("receive error:", err)
			continue
		}

		message := string(buf[:n])
		fmt.Printf("Server: %s\n", message)

		if message == "exit" {
			return
		}
	}
}
