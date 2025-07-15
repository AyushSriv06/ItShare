package connection

import (
	"ItShare/server/interfaces"
	"fmt"
	"net"
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




