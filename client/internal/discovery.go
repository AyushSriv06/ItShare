package connection

import (
	"bufio"
	"ItShare/utils"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// DiscoveredServer represents a discovered server
type DiscoveredServer struct {
	Address string
	IP      string
	Port    string
}

// DiscoverServers listens for UDP broadcast messages from DrizLink servers
func DiscoverServers(timeout time.Duration) ([]DiscoveredServer, error) {
	fmt.Println(utils.InfoColor("🔍 Scanning for DrizLink servers on local network..."))

	// Create UDP connection to listen for broadcasts
	addr, err := net.ResolveUDPAddr("udp", ":9876")
	if err != nil {
		return nil, fmt.Errorf("error resolving UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("error creating UDP listener: %v", err)
	}
	defer conn.Close()

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(timeout))

	var servers []DiscoveredServer
	serverMap := make(map[string]bool) // To avoid duplicates

	buffer := make([]byte, 1024)

	fmt.Println(utils.InfoColor("   Listening for broadcasts for"), utils.CommandColor(timeout.String()))

	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			// Check if it's a timeout
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			continue
		}

		message := string(buffer[:n])

		// Parse broadcast message: DRIZLINK_SERVER:<server_ip>:<server_port>
		if strings.HasPrefix(message, "DRIZLINK_SERVER:") {
			parts := strings.Split(message, ":")
			if len(parts) == 3 {
				serverIP := parts[1]
				serverPort := parts[2]
				serverAddress := fmt.Sprintf("%s:%s", serverIP, serverPort)

				// Avoid duplicates
				if !serverMap[serverAddress] {
					serverMap[serverAddress] = true
					servers = append(servers, DiscoveredServer{
						Address: serverAddress,
						IP:      serverIP,
						Port:    serverPort,
					})

					fmt.Printf("%s Found server: %s (from %s)\n",
						utils.SuccessColor("✅"),
						utils.InfoColor(serverAddress),
						utils.CommandColor(addr.IP.String()))
				}
			}
		}
	}

	return servers, nil
}

// SelectServer prompts user to select from discovered servers
func SelectServer(servers []DiscoveredServer) (string, error) {
	if len(servers) == 0 {
		return "", fmt.Errorf("no servers discovered")
	}

	fmt.Println(utils.HeaderColor("\n📡 Discovered DrizLink Servers:"))
	fmt.Println(utils.InfoColor("--------------------------------"))

	for i, server := range servers {
		fmt.Printf("%s %s %s\n",
			utils.CommandColor(fmt.Sprintf("[%d]", i+1)),
			utils.InfoColor("Server at"),
			utils.SuccessColor(server.Address))
	}

	fmt.Printf("%s %s\n",
		utils.CommandColor(fmt.Sprintf("[%d]", len(servers)+1)),
		utils.WarningColor("Enter server address manually"))

	fmt.Println(utils.InfoColor("--------------------------------"))
	fmt.Print(utils.CommandColor("Select server (1-" + fmt.Sprintf("%d", len(servers)+1) + "): "))

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil {
		return "", fmt.Errorf("invalid input")
	}

	if choice < 1 || choice > len(servers)+1 {
		return "", fmt.Errorf("invalid choice")
	}

	if choice == len(servers)+1 {
		return "", fmt.Errorf("manual_entry_requested")
	}

	selectedServer := servers[choice-1]
	fmt.Printf("%s Selected server: %s\n",
		utils.SuccessColor("✅"),
		utils.InfoColor(selectedServer.Address))

	return selectedServer.Address, nil
}
