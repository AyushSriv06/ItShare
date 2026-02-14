package connection

import (
	"ItShare/helper"
	"ItShare/server/interfaces"
	"ItShare/utils"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

func Connect(address string) (net.Listener, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return nil, err
	}
	return listener, nil
}

func Close(conn net.Conn) {
	conn.Close()
}

func Start(server *interfaces.Server) {
	listen, err := net.Listen("tcp", server.Address)
	if err != nil {
		fmt.Println("error in listen")
		panic(err)
	}

	defer listen.Close()

	// Initialize rooms map
	server.Rooms = make(map[string]*interfaces.Room)

	fmt.Println(utils.SuccessColor("✅ Server started on"), utils.InfoColor(server.Address))

	// Start UDP broadcast for server discovery
	go StartDiscoveryBroadcast(server.Address)

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("error in accept")
			continue
		}

		go HandleConnection(conn, server)
	}
}

func HandleConnection(conn net.Conn, server *interfaces.Server) {
	ipAddr := conn.RemoteAddr().String()
	ip := strings.Split(ipAddr, ":")[0]
	fmt.Println("New connection from", ip)
	/*
		// IP check disabled for local debugging
		if existingUser := server.IpAddresses[ip]; existingUser != nil {
			// ... code removed ...
		}
	*/

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("error in read username")
		return
	}
	username := string(buffer[:n])

	n, err = conn.Read(buffer)
	if err != nil {
		fmt.Println("error in read storeFilePath")
		return
	}
	storeFilePath := string(buffer[:n])

	userId := helper.GenerateUserId()

	user := &interfaces.User{
		UserId:        userId,
		Username:      username,
		StoreFilePath: storeFilePath,
		Conn:          conn,
		IsOnline:      true,
		IpAddress:     ip,
	}

	server.Mutex.Lock()
	server.Connections[user.UserId] = user
	server.IpAddresses[ip] = user
	server.Mutex.Unlock()

	// Send user ID to client
	conn.Write([]byte("/USERID " + userId + "\n"))

	welcomeMsg := fmt.Sprintf("User %s has joined the chat", username)
	BroadcastMessage(welcomeMsg, server, user)

	fmt.Printf("New user connected: %s (ID: %s)\n", username, userId)

	// Start handling messages for the new user
	handleUserMessages(conn, user, server)
}

// Room management functions
func CreateRoom(server *interfaces.Server, roomName string, creatorID string, memberIDs []string) (*interfaces.Room, error) {
	server.Mutex.Lock()
	defer server.Mutex.Unlock()

	// Generate unique room ID
	roomID := generateRoomID()

	// Create room
	room := &interfaces.Room{
		ID:        roomID,
		Name:      roomName,
		Members:   make(map[string]*interfaces.User),
		CreatedBy: creatorID,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	// Add creator to room
	if creator, exists := server.Connections[creatorID]; exists {
		room.Members[creatorID] = creator
	}

	// Add selected members to room
	for _, memberID := range memberIDs {
		if user, exists := server.Connections[memberID]; exists && user.IsOnline {
			room.Members[memberID] = user
		}
	}

	// Store room
	server.Rooms[roomID] = room

	return room, nil
}

func generateRoomID() string {
	return fmt.Sprintf("room_%d", rand.Intn(100000))
}

func AddUserToRoom(server *interfaces.Server, roomID string, userID string) error {
	server.Mutex.Lock()
	defer server.Mutex.Unlock()

	room, exists := server.Rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found")
	}

	user, exists := server.Connections[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	room.Mutex.Lock()
	room.Members[userID] = user
	room.Mutex.Unlock()

	return nil
}

func RemoveUserFromRoom(server *interfaces.Server, roomID string, userID string) error {
	server.Mutex.Lock()
	defer server.Mutex.Unlock()

	room, exists := server.Rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found")
	}

	room.Mutex.Lock()
	delete(room.Members, userID)
	room.Mutex.Unlock()

	return nil
}

func BroadcastRoomMessage(roomID string, senderUsername string, content string, server *interfaces.Server, sender *interfaces.User) {
	server.Mutex.Lock()
	room, exists := server.Rooms[roomID]
	server.Mutex.Unlock()

	if !exists {
		return
	}

	room.Mutex.RLock()
	defer room.Mutex.RUnlock()

	for _, member := range room.Members {
		if member.IsOnline && member != sender {
			_, _ = member.Conn.Write([]byte(fmt.Sprintf("[Room %s] %s: %s\n", room.Name, senderUsername, content)))
		}
	}
}

func GetOnlineUsersList(server *interfaces.Server) []interfaces.User {
	server.Mutex.Lock()
	defer server.Mutex.Unlock()

	var users []interfaces.User
	for _, user := range server.Connections {
		if user.IsOnline {
			users = append(users, *user)
		}
	}
	return users
}
func handleUserMessages(conn net.Conn, user *interfaces.User, server *interfaces.Server) {
	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("User disconnected: %s\n", user.Username)
			server.Mutex.Lock()
			user.IsOnline = false
			server.Mutex.Unlock()
			offlineMsg := fmt.Sprintf("User %s is now offline", user.Username)
			BroadcastMessage(offlineMsg, server, user)
			return
		}

		messageContent := string(buffer[:n])

		switch {
		case messageContent == "/exit":
			server.Mutex.Lock()
			user.IsOnline = false
			server.Mutex.Unlock()
			offlineMsg := fmt.Sprintf("User %s is now offline", user.Username)
			BroadcastMessage(offlineMsg, server, user)
			return
		case strings.HasPrefix(messageContent, "/FILE_REQUEST"):
			args := strings.SplitN(messageContent, " ", 6) // Updated to include checksum and transferID
			if len(args) < 4 {
				fmt.Println("Invalid arguments. Use: /FILE_REQUEST <userId> <filename> <fileSize> [checksum] [transferID]")
				continue
			}
			recipientId := args[1]
			fileName := args[2]
			fileSizeStr := strings.TrimSpace(args[3])
			fileSize, err := strconv.ParseInt(fileSizeStr, 10, 64)

			// Include checksum and transferID in filename if provided
			if len(args) >= 5 {
				checksum := strings.TrimSpace(args[4])
				fileName = fileName + "|" + checksum
				if len(args) == 6 {
					transferID := strings.TrimSpace(args[5])
					fileName = fileName + "|" + transferID
				}
			}

			if err != nil {
				fmt.Println("Invalid fileSize. Use: /FILE_REQUEST <userId> <filename> <fileSize> [checksum]")
				continue
			}

			HandleFileTransfer(server, conn, recipientId, fileName, fileSize)
			continue
		case strings.HasPrefix(messageContent, "/FOLDER_REQUEST"):
			args := strings.SplitN(messageContent, " ", 6) // Updated to include checksum and transferID
			if len(args) < 4 {
				fmt.Println("Invalid arguments. Use: /FOLDER_REQUEST <userId> <folderName> <folderSize> [checksum] [transferID]")
				continue
			}
			recipientId := args[1]
			folderName := args[2]
			folderSizeStr := strings.TrimSpace(args[3])
			folderSize, err := strconv.ParseInt(folderSizeStr, 10, 64)

			// Include checksum and transferID in foldername if provided
			if len(args) >= 5 {
				checksum := strings.TrimSpace(args[4])
				folderName = folderName + "|" + checksum
				if len(args) == 6 {
					transferID := strings.TrimSpace(args[5])
					folderName = folderName + "|" + transferID
				}
			}

			if err != nil {
				fmt.Println("Invalid folderSize. Use: /FOLDER_REQUEST <userId> <folderName> <folderSize> [checksum]")
				continue
			}

			HandleFolderTransfer(server, conn, recipientId, folderName, folderSize)
			continue
		case messageContent == "PONG\n":
			continue
		case strings.HasPrefix(messageContent, "/status"):
			_, err = conn.Write([]byte("USERS:"))
			if err != nil {
				fmt.Println("Error sending user list header:", err)
				continue
			}
			for _, user := range server.Connections {
				if user.IsOnline {
					statusMsg := fmt.Sprintf("%s (%s) is online\n", user.Username, user.UserId)
					_, err = conn.Write([]byte(statusMsg))
					if err != nil {
						fmt.Println("Error sending user list:", err)
						continue
					}
				}
			}
			continue
		case strings.HasPrefix(messageContent, "/GET_ONLINE_USERS"):
			users := GetOnlineUsersList(server)
			response := "ONLINE_USERS_LIST"
			for _, u := range users {
				if u.UserId != user.UserId { // Don't include the requesting user
					response += fmt.Sprintf(" %s|%s", u.UserId, u.Username)
				}
			}
			response += "\n"
			_, err = conn.Write([]byte(response))
			if err != nil {
				fmt.Println("Error sending online users list:", err)
			}
			continue
		case strings.HasPrefix(messageContent, "/CREATE_ROOM"):
			args := strings.SplitN(messageContent, " ", 3)
			if len(args) < 3 {
				fmt.Println("Invalid arguments. Use: /CREATE_ROOM <roomName> <userID1,userID2,...>")
				continue
			}
			roomName := args[1]
			memberIDsStr := args[2]
			memberIDs := strings.Split(memberIDsStr, ",")

			room, err := CreateRoom(server, roomName, user.UserId, memberIDs)
			if err != nil {
				fmt.Printf("Error creating room: %v\n", err)
				continue
			}

			// Notify all room members about room creation
			for memberID, member := range room.Members {
				if member.IsOnline {
					notification := fmt.Sprintf("ROOM_CREATED %s %s %s\n", room.ID, room.Name, user.Username)
					_, err = member.Conn.Write([]byte(notification))
					if err != nil {
						fmt.Printf("Error notifying user %s about room creation: %v\n", memberID, err)
					}
				}
			}
			continue
		case strings.HasPrefix(messageContent, "/JOIN_ROOM"):
			args := strings.SplitN(messageContent, " ", 2)
			if len(args) != 2 {
				fmt.Println("Invalid arguments. Use: /JOIN_ROOM <roomID>")
				continue
			}
			roomID := strings.TrimSpace(args[1])

			server.Mutex.Lock()
			room, exists := server.Rooms[roomID]
			server.Mutex.Unlock()

			if !exists {
				_, err = conn.Write([]byte("ROOM_NOT_FOUND\n"))
				if err != nil {
					fmt.Printf("Error sending room not found message: %v\n", err)
				}
				continue
			}

			// Check if user is a member of the room
			room.Mutex.RLock()
			_, isMember := room.Members[user.UserId]
			room.Mutex.RUnlock()

			if !isMember {
				_, err = conn.Write([]byte("NOT_ROOM_MEMBER\n"))
				if err != nil {
					fmt.Printf("Error sending not member message: %v\n", err)
				}
				continue
			}

			user.CurrentRoomID = roomID
			_, err = conn.Write([]byte(fmt.Sprintf("ROOM_JOINED %s %s\n", roomID, room.Name)))
			if err != nil {
				fmt.Printf("Error sending room joined confirmation: %v\n", err)
			}
			continue
		case strings.HasPrefix(messageContent, "/LEAVE_ROOM"):
			if user.CurrentRoomID != "" {
				oldRoomID := user.CurrentRoomID
				user.CurrentRoomID = ""
				_, err = conn.Write([]byte(fmt.Sprintf("ROOM_LEFT %s\n", oldRoomID)))
				if err != nil {
					fmt.Printf("Error sending room left confirmation: %v\n", err)
				}
			}
			continue
		case strings.HasPrefix(messageContent, "/LIST_ROOMS"):
			server.Mutex.Lock()
			response := "ROOMS_LIST"
			for roomID, room := range server.Rooms {
				room.Mutex.RLock()
				if _, isMember := room.Members[user.UserId]; isMember {
					response += fmt.Sprintf(" %s|%s|%d", roomID, room.Name, len(room.Members))
				}
				room.Mutex.RUnlock()
			}
			server.Mutex.Unlock()
			response += "\n"
			_, err = conn.Write([]byte(response))
			if err != nil {
				fmt.Printf("Error sending rooms list: %v\n", err)
			}
			continue
		case strings.HasPrefix(messageContent, "/ROOM_MESSAGE"):
			args := strings.SplitN(messageContent, " ", 3)
			if len(args) != 3 {
				fmt.Println("Invalid arguments. Use: /ROOM_MESSAGE <roomID> <content>")
				continue
			}
			roomID := args[1]
			content := args[2]

			// Verify user is member of the room
			server.Mutex.Lock()
			room, exists := server.Rooms[roomID]
			server.Mutex.Unlock()

			if !exists {
				continue
			}

			room.Mutex.RLock()
			_, isMember := room.Members[user.UserId]
			room.Mutex.RUnlock()

			if !isMember {
				continue
			}

			BroadcastRoomMessage(roomID, user.Username, content, server, user)
			continue
		case strings.HasPrefix(messageContent, "/LOOK"):
			args := strings.SplitN(messageContent, " ", 2)
			if len(args) != 2 {
				fmt.Println("Invalid arguments. Use: /LOOK <userId>")
				continue
			}
			recipientId := strings.TrimSpace(args[1])
			HandleLookupRequest(server, conn, recipientId)
			continue
		case strings.HasPrefix(messageContent, "/DIR_LISTING"):
			args := strings.SplitN(messageContent, " ", 3)
			if len(args) != 3 {
				fmt.Println("Invalid arguments. Use: /DIR_LISTING <userId> <files>")
				continue
			}
			userId := strings.TrimSpace(args[1])
			files := strings.TrimSpace(args[2])
			HandleLookupResponse(server, conn, userId, strings.Split(files, " "))
			continue
		case strings.HasPrefix(messageContent, "/DOWNLOAD_REQUEST"):
			args := strings.SplitN(messageContent, " ", 3)
			if len(args) != 3 {
				fmt.Println("Invalid arguments. Use: /DOWNLOAD_REQUEST <userId> <filename>")
				continue
			}
			senderId := strings.TrimSpace(args[1])
			recipientId := user.UserId
			filePath := strings.TrimSpace(args[2])
			HandleDownloadRequest(server, conn, senderId, recipientId, filePath)
			continue
		case strings.HasPrefix(messageContent, "/ROOM_MEMBERS"):
			args := strings.SplitN(messageContent, " ", 2)
			if len(args) != 2 {
				conn.Write([]byte("ROOM_MEMBERS_RESPONSE ERROR\n"))
				continue
			}
			roomID := strings.TrimSpace(args[1])
			server.Mutex.Lock()
			room, exists := server.Rooms[roomID]
			server.Mutex.Unlock()
			if !exists {
				conn.Write([]byte("ROOM_MEMBERS_RESPONSE ERROR\n"))
				continue
			}
			room.Mutex.RLock()
			var userIDs []string
			for uid := range room.Members {
				userIDs = append(userIDs, uid)
			}
			room.Mutex.RUnlock()
			response := "ROOM_MEMBERS_RESPONSE " + roomID + " " + strings.Join(userIDs, ",") + "\n"
			conn.Write([]byte(response))
			continue
		default:
			// Check if user is in a room and wants to send a room message
			if user.CurrentRoomID != "" {
				BroadcastRoomMessage(user.CurrentRoomID, user.Username, messageContent, server, user)
			} else {
				BroadcastMessage(messageContent, server, user)
			}
		}
	}
}

func BroadcastMessage(content string, server *interfaces.Server, sender *interfaces.User) {
	server.Mutex.Lock()
	defer server.Mutex.Unlock()
	for _, recipient := range server.Connections {
		if recipient.IsOnline && recipient != sender {
			_, _ = recipient.Conn.Write([]byte(fmt.Sprintf("%s: %s\n", sender.Username, content)))
		}
	}
}

func StartHeartBeat(interval time.Duration, server *interfaces.Server) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			server.Mutex.Lock()
			for _, user := range server.Connections {
				if user.IsOnline {
					_, err := user.Conn.Write([]byte("PING\n"))
					if err != nil {
						fmt.Printf("User disconnected: %s\n", user.Username)
						user.IsOnline = false
						BroadcastMessage(fmt.Sprintf("User %s is now offline", user.Username), server, user)
					}
				}
			}
			server.Mutex.Unlock()
		}
	}()
}

// StartDiscoveryBroadcast continuously broadcasts server presence via UDP
func StartDiscoveryBroadcast(serverAddress string) {
	// Extract port from server address
	port := strings.TrimPrefix(serverAddress, ":")

	// Get the server's actual IP address
	serverIP, err := getServerIP()
	if err != nil {
		fmt.Println(utils.ErrorColor("❌ Error getting server IP for broadcast:"), err)
		return
	}

	// Create UDP connection for broadcasting
	conn, err := net.Dial("udp", "255.255.255.255:9876")
	if err != nil {
		fmt.Println(utils.ErrorColor("❌ Error creating UDP broadcast connection:"), err)
		return
	}
	defer conn.Close()

	// Broadcast message format: DRIZLINK_SERVER:<server_ip>:<server_port>
	broadcastMsg := fmt.Sprintf("DRIZLINK_SERVER:%s:%s", serverIP, port)

	fmt.Println(utils.InfoColor("📡 Broadcasting server presence on UDP port 9876"))
	fmt.Println(utils.InfoColor("   Message:"), utils.CommandColor(broadcastMsg))

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		_, err := conn.Write([]byte(broadcastMsg))
		if err != nil {
			fmt.Println(utils.ErrorColor("❌ Error broadcasting server presence:"), err)
			continue
		}
	}
}

// getServerIP returns the server's local IP address
func getServerIP() (string, error) {
	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String(), nil
				}
			}
		}
	}

	return "", fmt.Errorf("no suitable network interface found")
}
