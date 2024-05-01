package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Peer struct {
	IP          string
	Port        int
	PeerID      int
	Successor   *Peer
	Predecessor *Peer
	FileMap     map[int]string // FileID to filename map
	mu          sync.Mutex
}

func hashString(input string) int {
	hash := sha256.Sum256([]byte(input))
	// Convert the first 4 bytes of the hash to an integer
	hashValue := 0
	for i := 0; i < 4; i++ {
		hashValue = (hashValue << 8) | int(hash[i])
	}
	return hashValue
}

func NewPeer(ip string, port int) *Peer {
	id := hashString(ip + ":" + strconv.Itoa(port))
	return &Peer{
		IP:          ip,
		Port:        port,
		PeerID:      id,
		Successor:   nil,
		Predecessor: nil,
		FileMap:     make(map[int]string),
	}
}

func (p *Peer) findSuccessor(fileID int) *Peer {
	current := p
	for {
		if fileID > current.PeerID && fileID <= current.Successor.PeerID {
			return current.Successor
		}
		current = current.Successor
		if current == p {
			return p
		}
	}
}

func (p *Peer) storeFile(filename string) {
	fileID := hashString(filename)
	successor := p.findSuccessor(fileID)
	successor.mu.Lock()
	defer successor.mu.Unlock()
	successor.FileMap[fileID] = filename
	// Log the filename and file ID
	fmt.Printf("File '%s' (ID: %d) stored on peer %d.\n", filename, fileID, successor.PeerID)
}

func (p *Peer) retrieveFile(fileID int) (string, bool) {
	successor := p.findSuccessor(fileID)
	successor.mu.Lock()
	defer successor.mu.Unlock()
	filename, exists := successor.FileMap[fileID]
	// Log the file ID and peer ID for debugging
	fmt.Printf("Retrieving file with ID: %d from peer %d.\n", fileID, successor.PeerID)
	return filename, exists
}

func (p *Peer) handleClientConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		request, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading request:", err)
			return
		}
		request = strings.TrimSpace(request)

		// Process the request
		parts := strings.Split(request, " ")
		command := parts[0]
		switch command {
		case "CONNECT":
			address := parts[1]
			ipPort := strings.Split(address, ":")
			ip := ipPort[0]
			port, _ := strconv.Atoi(ipPort[1])

			newPeer := NewPeer(ip, port)
			// Update the new peer's successor and predecessor
			newPeer.Predecessor = p.Predecessor
			newPeer.Successor = p
			p.Predecessor.Successor = newPeer
			p.Predecessor = newPeer

			// Notify the connecting peer
			writer.WriteString("OK\n")
			writer.Flush()
		case "STORE":
			filename := parts[1]
			p.storeFile(filename)
			writer.WriteString("File stored successfully.\n")
			writer.Flush()
		case "RETRIEVE":
			filename := parts[1]
			fileID := hashString(filename)
			file, exists := p.retrieveFile(fileID)
			if exists {
				writer.WriteString(fmt.Sprintf("File found: %s\n", file))
			} else {
				writer.WriteString("File not found.\n")
			}
			writer.Flush()
		case "DISPLAY":
			writer.WriteString(fmt.Sprintf("PeerID: %d, Successor: %d, Predecessor: %d\n",
				p.PeerID, p.Successor.PeerID, p.Predecessor.PeerID))
			writer.Flush()
		case "EXIT":
			return
		default:
			writer.WriteString("Unknown command.\n")
			writer.Flush()
		}
	}
}

func (p *Peer) StartServer() {
	addr := fmt.Sprintf(":%d", p.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Failed to start server:", err)
		os.Exit(1)
	}
	fmt.Println("Server started on port", p.Port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}
		go p.handleClientConnection(conn)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ./task2-peer <port>")
		os.Exit(1)
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Invalid port number:", os.Args[1])
		os.Exit(1)
	}

	peer := NewPeer("localhost", port)
	// Initially, the peer's successor and predecessor are itself
	peer.Successor = peer
	peer.Predecessor = peer

	peer.StartServer()
}
