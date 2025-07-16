package helper

import (
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

func CheckServerAvailability(address string) (bool, string) {
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		// Provide more specific error information
		if strings.Contains(err.Error(), "connection refused") {
			return false, "Connection refused - no server running at this address"
		} else if strings.Contains(err.Error(), "no such host") {
			return false, "Host not found - check if the hostname is correct"
		} else if strings.Contains(err.Error(), "i/o timeout") {
			return false, "Connection timed out - server might be behind a firewall"
		}
		return false, err.Error()
	}
	conn.Close()
	return true, ""
}

func GenerateUserId() string {
	return strconv.Itoa(rand.Intn(10000000))
}
