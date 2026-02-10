package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"bytes"
	"github.com/creack/pty"
)

const (
	MESSAGELIMIT = 4096
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
	fmt.Println("Starting doom-ascii stream...")

	if err := streamDoomOutput(conn, clientAddr); err != nil {
		fmt.Println("doom stream error:", err)
	}
}

func streamDoomOutput(conn *net.UDPConn, clientAddr *net.UDPAddr) error {
	wadFile := "DOOM1.wad"
	if _, err := os.Stat(filepath.Join("doom", wadFile)); err != nil {
		if _, upperErr := os.Stat(filepath.Join("doom", "DOOM1.WAD")); upperErr == nil {
			wadFile = "DOOM1.WAD"
		}
	}

	cmd := exec.Command("./doom-ascii", "-iwad", wadFile)
	cmd.Dir = "doom"

	ptyFile, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("start doom-ascii in pty: %w", err)
	}
	defer ptyFile.Close()

	waitCh := make(chan error, 1)
	go func() {
		waitErr := cmd.Wait()
		waitCh <- waitErr
		close(waitCh)
	}()

	buf := make([]byte, MESSAGELIMIT)
	for {
		n, err := ptyFile.Read(buf)
		if n > 0 {
			if _, writeErr := conn.WriteToUDP(buf[:n], clientAddr); writeErr != nil {
				return fmt.Errorf("send doom output: %w", writeErr)
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, syscall.EIO) {
				break
			}
			return fmt.Errorf("read doom output: %w", err)
		}
	}

	if _, err := conn.WriteToUDP([]byte("exit"), clientAddr); err != nil {
		return fmt.Errorf("send exit: %w", err)
	}

	if err := <-waitCh; err != nil {
		return fmt.Errorf("doom-ascii exited with error: %w", err)
	}

	return nil
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

		if n == 4 && string(buf[:n]) == "exit" {
			return
		}

		if _, err := os.Stdout.Write(buf[:n]); err != nil {
			fmt.Println("stdout write error:", err)
			return
		}
	}
}

func rlEncode(input []byte) []byte{
	if len(input) == 0{
		return nil
	}
	var buf bytes.Buffer
	cnt := 1
	for i := 1; i<len(input); i++{
		if input[i] == input[i-1] && cnt < 255{
			cnt++
		}else{
			buf.WriteByte(byte(cnt))
			buf.WriteByte(input[i-1])
			cnt = 1
		}
	}
	buf.WriteByte(byte(cnt))
	buf.WriteByte(input[len(input)-1])

	return buf.Bytes()
}

func rlDecode(input []byte) ([]byte, error){
	var buf bytes.Buffer
	for i := 0; i<len(input); i += 2{
		if i+1 >= len(input){
			return nil, fmt.Errorf("invalid rle data")
		}
		cnt := int(input[i])
		b:= input[i+1]
		
		for j:=0; j < cnt; j++{
			buf.WriteByte(b)
		}
	}
	return buf.Bytes(), nil
}
