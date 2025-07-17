package main

import (
	connection "ItShare/client/internal"
	"ItShare/helper"
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func promptForServerAddress() string {
	reader := bufio.NewReader(os.Stdin)
	
	for {
		address, _ := reader.ReadString('\n')
		address = strings.TrimSpace(address)
		
		if !strings.Contains(address, ":") {
			continue
		}
		
		// Check if server is available at this address
				available, errMsg := helper.CheckServerAvailability(address)
		if !available {
			fmt.Print(errMsg)
			retry, _ := reader.ReadString('\n')
			retry = strings.TrimSpace(strings.ToLower(retry))
			
			if retry != "y" && retry != "yes" {
				os.Exit(1)
			}
			continue
		}
		return address
	}
}

func main() {
	serverAddr := flag.String("server", "", "Server address in format host:port")
	flag.Parse()
	
	
	// If server address not provided via command line, ask user
	address := *serverAddr
	if address == "" {
		address = promptForServerAddress()
	} else {
		
		// Check if server is available
		available, errMsg := helper.CheckServerAvailability(address)
		if !available {
			fmt.Print(errMsg)
			return
		}
	}
	
	conn, err := connection.Connect(address)
		if err != nil {
		if err.Error() == "reconnect" {
			goto startChat 
		} else {
			return
		}
	}

	defer connection.Close(conn)

	err = connection.UserInput("Username", conn)
	if err != nil {
		if err.Error() == "reconnect" {
			goto startChat
		} else {
			fmt.Print(err)
			return
		}
	}


	err = connection.UserInput("Store File Path", conn)
	if err != nil {
		if err.Error() == "reconnect" {
			goto startChat
		} else {
			fmt.Print(err)
			return
		}
	}
	startChat:

	go connection.ReadLoop(conn)

}
