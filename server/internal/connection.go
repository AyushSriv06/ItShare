package connection

import (
	"ItShare/helper"
	"ItShare/server/interfaces"
	"fmt"
	"net"
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
	fmt.Println("Server started on", server.Address)

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
	if existingUser := server.IpAddresses[ip]; existingUser != nil {
		fmt.Println("Connection already exists for IP:", ip)
		// Send reconnection signal with existing user data
		reconnectMsg := fmt.Sprintf("/RECONNECT %s %s", existingUser.Username, existingUser.StoreFilePath)
		_, err := conn.Write([]byte(reconnectMsg))
		if err != nil {
			fmt.Println("Error sending reconnect signal:", err)
			return
		}

		// Update connection and online status
		server.Mutex.Lock()
		existingUser.Conn = conn
		existingUser.IsOnline = true
		server.Mutex.Unlock()

		// Encrypt and broadcast welcome back message
		welcomeMsg := fmt.Sprintf("User %s has rejoined the chat", existingUser.Username)
		BroadcastMessage(welcomeMsg, server, existingUser)

		// Start handling messages for the reconnected user
		return
	}

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

	welcomeMsg := fmt.Sprintf("User %s has joined the chat", username)
	BroadcastMessage(welcomeMsg, server, user)

	fmt.Printf("New user connected: %s (ID: %s)\n", username, userId)

	// Start handling messages for the new user
	
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





