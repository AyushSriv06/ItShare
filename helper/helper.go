package helper

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

//computes an MD5 hash of a file
func CalculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// computes an MD5 hash from a reader without consuming it
// Returns the checksum and a new reader that can be used normally
func CalculateDataChecksum(reader io.Reader) (string, io.Reader, error) {
	hash := md5.New()
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", nil, err
	}
	
	hash.Write(data)
	checksum := hex.EncodeToString(hash.Sum(nil))
	
	// Return a new reader with the same data
	return checksum, bytes.NewReader(data), nil
}

// checks if two checksums match
func VerifyChecksum(original, received string) bool {
	return original == received
}

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
