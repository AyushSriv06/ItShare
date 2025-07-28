package connection

import (
	"ItShare/helper"
	"ItShare/utils"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func HandleSendFile(conn net.Conn, recipientId, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(utils.ErrorColor("❌ Error opening file:"), err)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(utils.ErrorColor("❌ Error getting file info:"), err)
		return
	}

	fileSize := fileInfo.Size()
	fileName := fileInfo.Name()

	// Calculate checksum of file
	checksum, err := helper.CalculateFileChecksum(filePath)
	if err != nil {
		fmt.Println(utils.ErrorColor("❌ Error calculating checksum:"), err)
		return
	}

	transferID := GenerateTransferID()
	fmt.Printf("%s Sending file '%s' to user %s (Transfer ID: %s)...\n",
		utils.InfoColor("📤"),
		utils.InfoColor(fileName),
		utils.UserColor(recipientId),
		utils.CommandColor(transferID))

	// Send file request with file size, checksum, and transfer ID
	_, err = conn.Write([]byte(fmt.Sprintf("/FILE_REQUEST %s %s %d %s %s\n",
		recipientId, fileName, fileSize, checksum, transferID)))
	if err != nil {
		fmt.Println(utils.ErrorColor("❌ Error sending file request:"), err)
		return
	}

	// Create progress bar with transfer ID
	bar := utils.CreateProgressBar(fileSize, "📤 Sending file")
	bar.SetTransferId(transferID)

	transfer := &Transfer{
		ID:            transferID,
		Type:          FileTransfer,
		Name:          fileName,
		Size:          fileSize,
		BytesComplete: 0,
		Status:        Active,
		Direction:     "send",
		Recipient:     recipientId,
		Path:          filePath,
		Checksum:      checksum,
		StartTime:     time.Now(),
		File:          file,
		Connection:    conn,
		ProgressBar:   bar,
	}

	RegisterTransfer(transfer)

	reader := NewCheckpointedReader(file, transfer, 32768) // 32KB chunks

	n, err := io.CopyN(conn, io.TeeReader(reader, bar), fileSize)

	if err != nil {
		UpdateTransferStatus(transferID, Failed)
		fmt.Println(utils.ErrorColor("\n❌ Error sending file:"), err)
		RemoveTransfer(transferID)
		return
	}

	if n != fileSize {
		UpdateTransferStatus(transferID, Failed)
		fmt.Println(utils.ErrorColor("\n❌ Error: sent"), utils.ErrorColor(n),
			utils.ErrorColor("bytes, expected"), utils.ErrorColor(fileSize), utils.ErrorColor("bytes"))
		RemoveTransfer(transferID)
		return
	}

	// Mark transfer as completed
	UpdateTransferStatus(transferID, Completed)

	fmt.Printf("%s File '%s' sent successfully!\n",
		utils.SuccessColor("\n✅"),
		utils.SuccessColor(fileName))
	fmt.Println(utils.InfoColor("  MD5 Checksum:"), utils.InfoColor(checksum))

	// Clean up the transfer
	RemoveTransfer(transferID)
}

func HandleFileTransfer(conn net.Conn, recipientId, fileName string, fileSize int64, storeFilePath string) {
	// Get checksum and transfer ID from the split content
	parts := strings.SplitN(fileName, "|", 3)
	checksum := ""
	transferID := ""

	if len(parts) == 3 {
		fileName = parts[0]
		checksum = parts[1]
		transferID = parts[2]
	} else if len(parts) == 2 {
		fileName = parts[0]
		checksum = parts[1]
		transferID = GenerateTransferID()
	} else {
		transferID = GenerateTransferID()
	}

	fmt.Printf("[DEBUG] After split: fileName: %s, checksum: %s, transferID: %s\n", fileName, checksum, transferID)

	fmt.Printf("%s Receiving file: %s (Size: %s, Transfer ID: %s)\n",
		utils.InfoColor("📥"),
		utils.InfoColor(fileName),
		utils.InfoColor(fmt.Sprintf("%d bytes", fileSize)),
		utils.CommandColor(transferID))

	filePath := filepath.Join(storeFilePath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println(utils.ErrorColor("❌ Error creating file:"), err)
		return
	}
	defer file.Close()

	// Create progress bar with transfer ID
	bar := utils.CreateProgressBar(fileSize, "📥 Receiving file")
	bar.SetTransferId(transferID)

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
		ProgressBar:   bar,
	}

	RegisterTransfer(transfer)

	writer := NewCheckpointedWriter(file, transfer, 32768) // 32KB chunks

	// Write to file and update progress bar simultaneously
	n, err := io.CopyN(writer, io.TeeReader(conn, bar), fileSize)

	if err != nil {
		UpdateTransferStatus(transferID, Failed)
		fmt.Println(utils.ErrorColor("\n❌ Error receiving file:"), err)
		RemoveTransfer(transferID)
		return
	}

	if n != fileSize {
		UpdateTransferStatus(transferID, Failed)
		fmt.Println(utils.ErrorColor("\n❌ Error: received"), utils.ErrorColor(n),
			utils.ErrorColor("bytes, expected"), utils.ErrorColor(fileSize), utils.ErrorColor("bytes"))
		RemoveTransfer(transferID)
		return
	}

	// Verify checksum if provided
	if checksum != "" {
		file.Close() // Close file before calculating checksum
		receivedChecksum, err := helper.CalculateFileChecksum(filePath)
		if err != nil {
			fmt.Println(utils.ErrorColor("\n❌ Error calculating checksum:"), err)
		} else {
			fmt.Println(utils.InfoColor("\n📋 Calculated checksum:"), utils.InfoColor(receivedChecksum))

			if helper.VerifyChecksum(checksum, receivedChecksum) {
				fmt.Println(utils.SuccessColor("✅ Checksum verification successful! File integrity confirmed."))
			} else {
				fmt.Println(utils.ErrorColor("❌ Checksum verification failed! File may be corrupted."))
			}
		}
	}

	// Mark transfer as completed
	UpdateTransferStatus(transferID, Completed)

	fmt.Printf("%s File '%s' received successfully!\n",
		utils.SuccessColor("✅"),
		utils.SuccessColor(fileName))
	fmt.Println(utils.InfoColor("📂 Saved to:"), utils.InfoColor(filePath))

	// Clean up the transfer
	RemoveTransfer(transferID)
}

func HandleDownloadRequest(conn net.Conn, recipientId, filePath string) {
	_, err := conn.Write([]byte(fmt.Sprintf("/DOWNLOAD_REQUEST %s %s\n", recipientId, filePath)))
	if err != nil {
		fmt.Println("Error sending file request:", err)
		return
	}
	fmt.Println("File download request sent successfully")
}

func HandleDownloadResponse(conn net.Conn, userId, filePath string) {
	cleanPath := filepath.Clean(strings.TrimSpace(filePath))
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		fmt.Printf("Error resolving absolute path: %v\n", err)
		return
	}

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		fmt.Println("error in stat file", err)
		return
	}
	if !fileInfo.IsDir() {
		HandleSendFile(conn, userId, absPath)
	} else {
		HandleSendFolder(conn, userId,absPath)
	}
}