package connection

import (
	"ItShare/server/interfaces"
	"fmt"
	"io"
	"net"
	"strings"
)

func HandleFileTransfer(server *interfaces.Server, conn net.Conn, recipientId, fileName string, fileSize int64) {
	// Extract checksum and transfer ID if present
	checksum := ""
	transferID := ""
	fileNameWithChecksum := fileName

	parts := strings.SplitN(fileName, "|", 2)
	if len(parts) == 2 {
		fileName = parts[0]
		checksum = parts[1]
	}

	if checksum != "" && transferID != "" {
		fileNameWithChecksum = fmt.Sprintf("%s|%s|%s", fileName, checksum, transferID)
	} else if checksum != "" {
		fileNameWithChecksum = fmt.Sprintf("%s|%s", fileName, checksum)
	}

	recipient, exists := server.Connections[recipientId]
	if exists {
		// Include checksum in response if available
		_, err := recipient.Conn.Write([]byte(fmt.Sprintf("/FILE_RESPONSE\t%s\t%s\t%d\t%s\n",
			recipientId, fileNameWithChecksum, fileSize, recipient.StoreFilePath)))
		if err != nil {
			fmt.Printf("Error sending file response to %s: %v\n", recipientId, err)
		}
		n, err := io.CopyN(recipient.Conn, conn, fileSize)
		if err != nil {
			fmt.Printf("Error receiving file from %s: %v\n", recipientId, err)
		}
		fmt.Printf("Transferred %d bytes from %s\n", n, recipientId)
		if err != nil {
			fmt.Printf("Error sending file to %s: %v\n", recipientId, err)
		}
	} else {
		fmt.Printf("User %s not found\n", recipientId)
	}
}

func SendFile(server *interfaces.Server, senderId, recipientId, filePath string) {
	server.Mutex.Lock()
	defer server.Mutex.Unlock()

	_, exists := server.Connections[recipientId]
	if !exists {
		fmt.Printf("User %s not found\n", recipientId)
		return
	}

	sender, exists := server.Connections[senderId]
	if !exists {
		fmt.Printf("User %s not found\n", senderId)
		return
	}

	_, err := sender.Conn.Write([]byte(fmt.Sprintf("/sendfile %s %s\n", recipientId, filePath)))
	if err != nil {
		fmt.Printf("Error sending file to %s: %v\n", recipientId, err)
	}
}

func HandleDownloadRequest(server *interfaces.Server, conn net.Conn, senderId, recipientId, filePath string) {
	sender, exists := server.Connections[senderId]
	if !exists {
		fmt.Printf("User %s not found\n", senderId)
		return
	}

	if !sender.IsOnline {
		fmt.Printf("User %s is not online\n", senderId)
		return
	}

	_, err := sender.Conn.Write([]byte(fmt.Sprintf("/DOWNLOAD_REQUEST %s %s\n", recipientId, filePath)))
	if err != nil {
		fmt.Printf("Error sending file request to %s: %v\n", senderId, err)
	}
	fmt.Println("Download request sent successfully")
}
