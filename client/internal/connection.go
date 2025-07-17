package connection

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
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

func ReadLoop(conn net.Conn) {
	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Print( err)
			return
		}
		message := string(buffer[:n])
		switch {
		case strings.HasPrefix(message, "/FILE_RESPONSE"):
			args := strings.SplitN(message, " ", 5)
			if len(args) != 5 {
				continue
			}
			recipientId := args[1]
			fileName := args[2]
			fileSizeStr := strings.TrimSpace(args[3])
			fileSize, err := strconv.ParseInt(fileSizeStr, 10, 64)
			storeFilePath := args[4]
			if err != nil {
				continue
			}

			HandleFileTransfer(conn, recipientId, fileName, int64(fileSize), storeFilePath)
			continue
		case strings.HasPrefix(message, "/FOLDER_RESPONSE"):
			args := strings.SplitN(message, " ", 5)
			if len(args) != 5 {
				continue
			}
			
			HandleFolderTransfer(conn, recipientId, folderName, folderSize, storeFilePath)
			continue
		case strings.HasPrefix(message, "PING"):
			_, err = conn.Write([]byte("PONG\n"))
			if err != nil {
				fmt.Println(err)
				continue
			}
		case message == "USERS:":

			// Read the complete user list with timeout
			userList := ""
			tempBuf := make([]byte, 1024)
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))

			for {
				m, err := conn.Read(tempBuf)
				if err != nil {
					break // Break on error (likely timeout)
				}
				userList += string(tempBuf[:m])
				if m < 1024 {
					break // All data received
				}
			}

			// Reset the deadline
			conn.SetReadDeadline(time.Time{})

			// Process users
			userCount := 0
			for _, line := range strings.Split(userList, "\n") {
				if strings.TrimSpace(line) != "" {
					userCount++
					// Enhanced formatting for username and ID
					if strings.Contains(line, "[ID:") {
						parts := strings.SplitN(line, "[ID:", 2)
						if len(parts) == 2 {
							username := strings.TrimSpace(parts[0])
							idPart := strings.SplitN(parts[1], "]", 2)
							if len(idPart) == 2 {
								userId := strings.TrimSpace(idPart[0])
								status := strings.TrimSpace(idPart[1])
								continue
							}
						}
					}
				}
			}
		}
	}
}
