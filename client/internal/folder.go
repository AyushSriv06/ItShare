package connection

import (
	"fmt"
	"io"
	"net"
	"path/filepath"
	"ItShare/helper"
	"os"
	"strings"
	"time"
)

func HandleFolderTransfer(conn net.Conn, recipientId, folderName string, folderSize int64, storeFilePath string) {
	// Extract checksum and transfer ID if present
	checksum := ""
	transferID := ""

	parts := strings.SplitN(folderName, "|", 3)
	if len(parts) >= 2 {
		folderName = parts[0]
		checksum = parts[1]
		if len(parts) >= 3 {
			transferID = parts[2]
		} else {
			transferID = GenerateTransferID()
		}
	} else {
		transferID = GenerateTransferID()
	}


	// Create temporary zip file to store received data
	tempZipPath := filepath.Join(storeFilePath, folderName+".zip")
	zipFile, err := os.Create(tempZipPath)
	if err != nil {
		fmt.Println( err)
		return
	}


	// Create transfer record
	transfer := &Transfer{
		ID:            transferID,
		Type:          FolderTransfer,
		Name:          folderName,
		Size:          folderSize,
		BytesComplete: 0,
		Status:        Active,
		Direction:     "receive",
		Recipient:     recipientId,
		Path:          tempZipPath,
		Checksum:      checksum,
		StartTime:     time.Now(),
		File:          zipFile,
		Connection:    conn,
	}

	RegisterTransfer(transfer)

	writer := NewCheckpointedWriter(zipFile, transfer, 32768) // 32KB chunks

	// Receive the zip file data with progress
	n, err := io.CopyN(writer, io.TeeReader(conn, bar), folderSize)
	zipFile.Close()

	if err != nil {
		UpdateTransferStatus(transferID, Failed)
		os.Remove(tempZipPath)
		fmt.Println(err)
		RemoveTransfer(transferID)
		return
	}

	if n != folderSize {
		UpdateTransferStatus(transferID, Failed)
		os.Remove(tempZipPath)
		RemoveTransfer(transferID)
		return
	}

	// Verify checksum if provided
	if checksum != "" {
		receivedChecksum, err := helper.CalculateFileChecksum(tempZipPath)
		if err != nil {
			fmt.Println(err)
		} else {

			if helper.VerifyChecksum(checksum, receivedChecksum) {
				fmt.Println("✅ Checksum verification successful! Folder integrity confirmed.")
			} else {
				fmt.Println("❌ Checksum verification failed! Folder may be corrupted.")
			}
		}
	}

	fmt.Println("\n📦 Extracting folder...")
	//Extract the zip file
	destPath := filepath.Join(storeFilePath, folderName)
	err = helper.ExtractZip(tempZipPath, destPath)
	if err != nil {
		UpdateTransferStatus(transferID, Failed)
		os.Remove(tempZipPath)
		fmt.Println(err)
		RemoveTransfer(transferID)
		return
	}

	UpdateTransferStatus(transferID, Completed)

	// Clean up the temporary zip file
	os.Remove(tempZipPath)

	RemoveTransfer(transferID)
}