package main

import (
	"ItShare/server/interfaces"
	connection "ItShare/server/internal"
	"flag"
	"strings"
)

func main() {
	port := flag.String("port", "8080", "The port to listen on")
	flag.Parse()

	formattedPort := *port
	if !strings.HasPrefix(formattedPort, ":") {
		formattedPort = ":" + formattedPort
	}

	server := interfaces.Server{
		Address:     formattedPort,
		Connections: make(map[string]*interfaces.User),
		IpAddresses: make(map[string]*interfaces.User),
		Messages:    make(chan interfaces.Message),
	}

	connection.Start(&server)
}
