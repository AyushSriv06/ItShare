package connection

import (
	"ItShare/helper"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
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
		fmt.Println(err)
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

func HandleLookupRequest(conn net.Conn, userId string) {
	_, err := conn.Write([]byte(fmt.Sprintf("/LOOK %s\n", userId)))
	if err != nil {
		fmt.Printf("Error sending look request: %v\n", err)
		return
	}
}

func HandleLookupResponse(conn net.Conn, storeFilePath string, userId string) {
	// Clean and normalize the path
	cleanPath := filepath.Clean(strings.TrimSpace(storeFilePath))
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		fmt.Printf("Error resolving absolute path: %v\n", err)
		return
	}

	// Verify directory exists and is accessible
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Store directory does not exist: %s\n", absPath)
		} else {
			fmt.Printf("Error accessing directory: %v\n", err)
		}
		return
	}

	if !info.IsDir() {
		fmt.Printf("Path is not a directory: %s\n", absPath)
		return
	}

	var folders []string
	var files []string

	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return nil
		}
		if path == absPath {
			return nil
		}

		// Get clean relative path
		absolutePath := filepath.ToSlash(path)

		if info.IsDir() {
			folders = append(folders, fmt.Sprintf("[FOLDER] %s (Size: %d bytes)", absolutePath, info.Size()))
		} else {
			files = append(files, fmt.Sprintf("[FILE] %s (Size: %d bytes)", absolutePath, info.Size()))
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		return
	}

	var allEntries []string
	if len(folders) > 0 {
		allEntries = append(allEntries, "=== FOLDERS ===")
		allEntries = append(allEntries, folders...)
	}
	if len(files) > 0 {
		if len(allEntries) > 0 {
			allEntries = append(allEntries, "") // Add spacing between folders and files
		}
		allEntries = append(allEntries, "=== FILES ===")
		allEntries = append(allEntries, files...)
	}

	if len(allEntries) == 0 {
		allEntries = append(allEntries, "Directory is empty")
	}

	response := fmt.Sprintf("LOOK_RESPONSE %s %s\n", userId, strings.Join(allEntries, "\n"))
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Printf("Error sending lookup response: %v\n", err)
	}

	for _, entry := range allEntries {
		fmt.Println(entry)
	}
}
