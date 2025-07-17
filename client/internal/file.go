package connection

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"ItShare/helper"
	"strings"
	"time"
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


	transfer := &Transfer{
		ID:            transferID,
		Type:          FileTransfer,
		Name:          fileName,
		Size:          fileSize,
		BytesComplete: 0,
		Status:        Active,
		Direction:     "receive",
		Recipient:     recipientId,
		Path:          filePath,
		Checksum:      checksum,
		StartTime:     time.Now(),
		File:          file,
		Connection:    conn,
	}

	RegisterTransfer(transfer)

	writer := NewCheckpointedWriter(file, transfer, 32768) // 32KB chunks

	// Write to file and update progress bar simultaneously
	n, err := io.CopyN(writer, io.TeeReader(conn,bar), fileSize)
	
	// bar will be defined late, it is the status bar

	if err != nil {
		UpdateTransferStatus(transferID, Failed)
		fmt.Println( err)
		RemoveTransfer(transferID)
		return
	}

	if n != fileSize {
		UpdateTransferStatus(transferID, Failed)
		fmt.Println("\n❌ Error: received")
		RemoveTransfer(transferID)
		return
	}

	// Verify checksum if provided
	if checksum != "" {
		file.Close() // Close file before calculating checksum
		receivedChecksum, err := helper.CalculateFileChecksum(filePath)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(receivedChecksum)

		}
	}

	// Mark transfer as completed
	UpdateTransferStatus(transferID, Completed)

	fmt.Printf("%s File '%s' received successfully!\n",)

	// Clean up the transfer
	RemoveTransfer(transferID)
}
