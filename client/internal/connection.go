package connection

import (
	"bufio"
	"fmt"
	"net"
	"errors"
	"os"
	"strings"
	"time"
)

func Connect(address string) (net.Conn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func Close(conn net.Conn) {
	conn.Close()
}

func UserInput(attribute string, conn net.Conn) error {
	// First check if we get a reconnection signal
	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buffer)
	conn.SetReadDeadline(time.Time{}) // Reset read deadline

	if err == nil && n > 0 {
		message := string(buffer[:n])
		if strings.HasPrefix(message, "/RECONNECT") {
			parts := strings.SplitN(message, " ", 4)
			if len(parts) == 3 {
				fmt.Printf("Welcome back %s!\n", parts[1])
				return errors.New("reconnect")
			}
		}
	}

	// If no reconnection signal, proceed with normal user input
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Enter your " + attribute + ": ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	// If it's a store file path, validate it
	if attribute == "Store File Path" {
		for {
			// Check if path exists
			if _, err := os.Stat(input); os.IsNotExist(err) {
				fmt.Println("Enter a valid " + attribute + ": ")
				input, _ = reader.ReadString('\n')
				input = strings.TrimSpace(input)
				continue
			}

			// Check if it's a directory
			fileInfo, err := os.Stat(input)
			if err != nil || !fileInfo.IsDir() {
				fmt.Println("Enter a valid " + attribute + ": ")
				input, _ = reader.ReadString('\n')
				input = strings.TrimSpace(input)
				continue
			}

			break
		}
	}

	_, err = conn.Write([]byte(input))
	if err != nil {
		fmt.Println("error in write " + attribute)
		panic(err)
	}

	return nil
}