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
	"ItShare/utils"
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
				fmt.Println(utils.ErrorColor("❌ Error: Directory does not exist")) 
				fmt.Println("Enter a valid " + attribute + ": ")
				input, _ = reader.ReadString('\n')
				input = strings.TrimSpace(input)
				continue
			}

			// Check if it's a directory
			fileInfo, err := os.Stat(input)
			if err != nil || !fileInfo.IsDir() {
				fmt.Println(utils.ErrorColor("❌ Error: Path is not a directory"))
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
			fmt.Println(utils.ErrorColor("❌ Connection lost:"), err)
			return
		}
		message := string(buffer[:n])
		switch {
		case strings.HasPrefix(message, "/FILE_RESPONSE"):
			fmt.Println(utils.InfoColor("📥 File transfer starting..."))
			args := strings.SplitN(message, " ", 5)
			if len(args) != 5 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /FILE_RESPONSE <userId> <filename> <fileSize> <storeFilePath>"))
				continue
			}
			recipientId := args[1]
			fileName := args[2]
			fileSizeStr := strings.TrimSpace(args[3])
			fileSize, err := strconv.ParseInt(fileSizeStr, 10, 64)
			storeFilePath := args[4]
			if err != nil {
				fmt.Println(utils.ErrorColor("❌ Invalid fileSize. Use: /FILE_RESPONSE <userId> <filename> <fileSize> <storeFilePath>"))
				continue
			}

			HandleFileTransfer(conn, recipientId, fileName, int64(fileSize), storeFilePath)
			continue
		case strings.HasPrefix(message, "/FOLDER_RESPONSE"):
			fmt.Println(utils.InfoColor("📥 Folder transfer starting..."))
			args := strings.SplitN(message, " ", 5)
			if len(args) != 5 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /FOLDER_RESPONSE <userId> <folderName> <folderSize> <storeFilePath>"))
				continue
			}
			
			recipientId := args[1]
			folderName := args[2]
			folderSizeStr := strings.TrimSpace(args[3])
			folderSize, err := strconv.ParseInt(folderSizeStr, 10, 64)
			storeFilePath := args[4]
			if err != nil {
				fmt.Println(utils.ErrorColor("❌ Invalid folderSize. Use: /FOLDER_RESPONSE <userId> <folderName> <folderSize> <storeFilePath>"))
				continue
			}
			HandleFolderTransfer(conn, recipientId, folderName, folderSize, storeFilePath)
			continue
		case strings.HasPrefix(message, "PING"):
			_, err = conn.Write([]byte("PONG\n"))
			if err != nil {
				fmt.Println(utils.ErrorColor("❌ Error responding to heartbeat:"), err)
				continue
			}
		case message == "USERS:":

			// Improved approach to accumulate the complete user list
			fmt.Println(utils.HeaderColor("\n👥 Online Users:"))
			fmt.Println(utils.InfoColor("-------------------"))

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
								fmt.Printf("%s %s %s %s %s\n",
									utils.SuccessColor(" •"),
									utils.UserColor(username),
									utils.InfoColor("(ID:"),
									utils.CommandColor(userId),
									utils.InfoColor(")"+status))
								continue
							}
						}
					}
				}
			}
			if userCount == 0 {
				fmt.Println(utils.InfoColor(" No users currently online"))
			}

			fmt.Println(utils.InfoColor("-------------------"))
			continue
		case strings.HasPrefix(message, "/LOOK_REQUEST"):
			args := strings.SplitN(message, " ", 3)
			if len(args) != 3 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /LOOK_REQUEST <storageFilePath> <userId>"))
				continue
			}
			storageFilePath := args[2]
			userId := args[1]
			fmt.Println(utils.InfoColor("🔍 Processing directory lookup request from"), utils.UserColor(userId))
			HandleLookupResponse(conn, storageFilePath, userId)
			continue
		case strings.HasPrefix(message, "/LOOK_RESPONSE"):
			args := strings.SplitN(message, " ", 3)
			if len(args) != 3 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /LOOK_RESPONSE <userId> <files>"))
				continue
			}
			userId := args[1]
			files := strings.Split(args[2], " ")

			fmt.Println(utils.HeaderColor("\n📂 Directory Listing for User:"), utils.UserColor(userId))
			fmt.Println(utils.InfoColor("-------------------------------------------"))

			for _, file := range files {
				if strings.HasPrefix(file, "[FOLDER]") {
					fmt.Println(utils.WarningColor("📁"), utils.InfoColor(file))
				} else if strings.HasPrefix(file, "[FILE]") {
					fmt.Println(utils.SuccessColor("📄"), utils.InfoColor(file))
				} else if strings.HasPrefix(file, "===") {
					fmt.Println(utils.HeaderColor(file))
				} else {
					fmt.Println(utils.InfoColor(file))
				}
			}

			fmt.Println(utils.InfoColor("-------------------------------------------\n"))
			continue
		case strings.HasPrefix(message, "/DOWNLOAD_REQUEST"):
			args := strings.SplitN(message, " ", 3)
			if len(args) != 3 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /DOWNLOAD_REQUEST <userId> <filename>"))
				continue
			}
			userId := args[1]
			filePath := args[2]
			fmt.Println(utils.InfoColor("📤 Download request from"), utils.UserColor(userId), utils.InfoColor("for"), utils.InfoColor(filePath))
			HandleDownloadResponse(conn, userId, filePath)
			continue
		default:
			if strings.Contains(message, "has joined the chat") {
				fmt.Println(utils.WarningColor("👋 " + message))
			} else if strings.Contains(message, "has rejoined the chat") {
				fmt.Println(utils.WarningColor("🔄 " + message))
			} else if strings.Contains(message, "is now offline") {
				fmt.Println(utils.WarningColor("👋 " + message))
			} else {
				fmt.Println(message)
			}
		}
	}
}

func WriteLoop(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(utils.CommandColor(">>> "))
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)
		switch {
		case message == "exit":
			fmt.Println(utils.InfoColor("👋 Goodbye!"))
			conn.Close()
			return
		case message == "/help":
			utils.PrintHelp()
			continue
		case strings.HasPrefix(message, "/sendfile"):
			args := strings.SplitN(message, " ", 3)
			if len(args) != 3 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /sendfile <userId> <filename>"))
				continue
			}
			recipientId := args[1]
			filePath := args[2]
			fmt.Println(utils.InfoColor("📤 Sending file to"), utils.UserColor(recipientId))
			HandleSendFile(conn, recipientId, filePath)
			continue
		case strings.HasPrefix(message, "/sendfolder"):
			args := strings.SplitN(message, " ", 3)
			if len(args) != 3 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /sendfolder <userId> <folderPath>"))
				continue
			}
			recipientId := args[1]
			folderPath := args[2]
			fmt.Println(utils.InfoColor("📤 Sending folder to"), utils.UserColor(recipientId))
			HandleSendFolder(conn, recipientId, folderPath)
			continue
		case strings.HasPrefix(message, "/lookup"):
			args := strings.SplitN(message, " ", 2)
			if len(args) != 2 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /lookup <userId>"))
				continue
			}
			recipientId := args[1]
			fmt.Println(utils.InfoColor("🔍 Looking up files for user"), utils.UserColor(recipientId))
			HandleLookupRequest(conn, recipientId)
			continue
		case strings.HasPrefix(message, "/status"):
			fmt.Println(utils.InfoColor("👥 Fetching online users..."))
			_, err := conn.Write([]byte(message))
			if err != nil {
				fmt.Println(utils.ErrorColor("❌ Error checking status:"), err)
				continue
			}
			continue
		case strings.HasPrefix(message, "/download"):
			args := strings.SplitN(message, " ", 3)
			if len(args) != 3 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /download <userId> <filename>"))
				continue
			}
			recipientId := args[1]
			filePath := args[2]
			fmt.Println(utils.InfoColor("📥 Requesting download from"), utils.UserColor(recipientId))
			HandleDownloadRequest(conn, recipientId, filePath)
			continue
		case strings.HasPrefix(message, "/transfers"):
			HandleListTransfers()
			continue
		case strings.HasPrefix(message, "/pause"):
			args := strings.SplitN(message, " ", 2)
			if len(args) != 2 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /pause <transferId>"))
				continue
			}
			transferID := args[1]
			HandlePauseTransfer(transferID)
			continue
		case strings.HasPrefix(message, "/resume"):
			args := strings.SplitN(message, " ", 2)
			if len(args) != 2 {
				fmt.Println(utils.ErrorColor("❌ Invalid arguments. Use: /resume <transferId>"))
				continue
			}
			transferID := args[1]
			HandleResumeTransfer(transferID)
			continue
		default:
			if message != "" {
				_, err := conn.Write([]byte(message))
				if err != nil {
					fmt.Println(utils.ErrorColor("❌ Error sending message:"), err)
					return
				}
			}
		}
	}
}