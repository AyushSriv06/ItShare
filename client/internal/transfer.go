package connection

import (
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

// TransferType represents the type of transfer
type TransferType int

const (
	FileTransfer TransferType = iota
	FolderTransfer
)

// TransferStatus represents the status of a transfer
type TransferStatus int

const (
	Active TransferStatus = iota
	Paused
	Completed
	Failed
)

// String representation of TransferStatus
func (s TransferStatus) String() string {
	switch s {
	case Active:
		return "Active"
	case Paused:
		return "Paused"
	case Completed:
		return "Completed"
	case Failed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// Transfer represents an active file or folder transfer
type Transfer struct {
	ID            string
	Type          TransferType
	Name          string
	Size          int64
	BytesComplete int64
	Status        TransferStatus
	Direction     string // "send" or "receive"
	Recipient     string
	Path          string
	Checksum      string
	StartTime     time.Time
	File          *os.File
	Connection    net.Conn
	PauseLock     sync.Mutex
	IsPaused      bool
}

// ActiveTransfers tracks all ongoing transfers
var (
	ActiveTransfers   = make(map[string]*Transfer)
	TransfersMutex    sync.RWMutex
	transferIDCounter = 1
)
func GenerateTransferID() string {
	TransfersMutex.Lock()
	defer TransfersMutex.Unlock()
	id := strconv.Itoa(transferIDCounter)
	transferIDCounter++
	return id
}