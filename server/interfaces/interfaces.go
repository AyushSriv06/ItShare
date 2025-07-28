package interfaces

import (
	"net"
	"sync"
)

type Server struct {
	Address     string
	Connections map[string]*User
	IpAddresses map[string]*User
	Messages    chan Message
	Rooms       map[string]*Room
	Mutex       sync.Mutex
}

type Message struct {
	SenderId       string
	SenderUsername string
	Content        string
	Timestamp      string
	RoomID         string
}

type User struct {
	UserId        string
	Username      string
	StoreFilePath string
	Conn          net.Conn
	IsOnline      bool
	IpAddress     string
	CurrentRoomID string
}

type Room struct {
	ID          string
	Name        string
	Members     map[string]*User
	CreatedBy   string
	CreatedAt   string
	Mutex       sync.RWMutex
}