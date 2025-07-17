package connection

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func HandleFileTransfer(conn net.Conn, recipientId, fileName string, fileSize int64, storeFilePath string) {
	// Get checksum and transfer ID from the split content
	parts := strings.SplitN(fileName, "|", 3)
	transferID := ""
	checksum := ""

	if len(parts) >= 2 {
		fileName = parts[0]
		checksum = parts[1]

		if len(parts) >= 3 {
			transferID = parts[2]
		} else {
			transferID = GenerateTransferID()
		}
	} else {
		transferID = GenerateTransferID()
	}
	filePath := filepath.Join(storeFilePath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println( err)
		return
	}
	defer file.Close()

	
}