package connection

import "net"

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