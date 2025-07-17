package helper

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"path/filepath"
	"encoding/hex"
	"fmt"
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

// ExtractZip extracts a zip archive to the specified destination
func ExtractZip(zipPath string, destPath string) error {
	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer archive.Close()

	for _, file := range archive.File {
		filePath := filepath.Join(destPath, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		srcFile, err := file.Open()
		if err != nil {
			dstFile.Close()
			return err
		}

		_, err = io.Copy(dstFile, srcFile)
		srcFile.Close()
		dstFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}