package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Client struct {
	Conn net.Conn
}

func NewClient(ip string, port int) (*Client, error) {
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &Client{Conn: conn}, nil
}

func (c *Client) storeFile(filename string) {
	fmt.Fprintf(c.Conn, "STORE %s\n", filename)
	reader := bufio.NewReader(c.Conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}
	fmt.Print(response)
}

func (c *Client) retrieveFile(filename string) {
	fmt.Fprintf(c.Conn, "RETRIEVE %s\n", filename)
	reader := bufio.NewReader(c.Conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}
	fmt.Print(response)
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: ./task2-client <ip> <port>")
		os.Exit(1)
	}

	ip := os.Args[1]
	port, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Invalid port number:", os.Args[2])
		os.Exit(1)
	}

	client, err := NewClient(ip, port)
	if err != nil {
		fmt.Println("Failed to connect to peer:", err)
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Please select an option:")
		fmt.Println("1) Enter the filename to store:")
		fmt.Println("2) Enter the filename to retrieve:")
		fmt.Println("3) Exit:")
		fmt.Print("Please select an option: ")

		choice, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			fmt.Print("Enter the filename to store: ")
			filename, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				continue
			}
			filename = strings.TrimSpace(filename)
			client.storeFile(filename)
		case "2":
			fmt.Print("Enter the filename to retrieve: ")
			filename, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				continue
			}
			filename = strings.TrimSpace(filename)
			client.retrieveFile(filename)
		case "3":
			fmt.Println("Exiting.")
			os.Exit(0)
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}
