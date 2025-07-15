package connection

import (
	"ItShare/server/interfaces"
	"fmt"
	"net"
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





