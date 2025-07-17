package main

import (
	"ItShare/helper"
	"ItShare/server/interfaces"
	connection "ItShare/server/internal"
	"ItShare/utils"
	"flag"
	"fmt"
	"strings"
	"time"
)

func main() {
	port := flag.String("port", "8080", "The port to listen on")
	flag.Parse()

	formattedPort := *port
	if !strings.HasPrefix(formattedPort, ":") {
		formattedPort = ":" + formattedPort
	}

	if helper.IsPortInUse(*port) {
		fmt.Println(utils.ErrorColor("❌ Error: Port " + *port + " is already in use"))
		fmt.Println(utils.InfoColor("Please choose a different port or stop the other server."))
		return
	}

	utils.PrintBanner()
	fmt.Println(utils.InfoColor("Starting server on port " + *port + "..."))

	server := interfaces.Server{
		Address:     formattedPort,
		Connections: make(map[string]*interfaces.User),
		IpAddresses: make(map[string]*interfaces.User),
		Messages:    make(chan interfaces.Message),
	}
	go connection.StartHeartBeat(100*time.Second, &server)
	connection.Start(&server)
}
