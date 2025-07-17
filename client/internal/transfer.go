package connection

import (
	"io"
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

// RegisterTransfer adds a new transfer to the tracking system
func RegisterTransfer(transfer *Transfer) {
	TransfersMutex.Lock()
	defer TransfersMutex.Unlock()
	ActiveTransfers[transfer.ID] = transfer
}


func GenerateTransferID() string {
	TransfersMutex.Lock()
	defer TransfersMutex.Unlock()
	id := strconv.Itoa(transferIDCounter)
	transferIDCounter++
	return id
}

// GetTransfer retrieves a transfer by ID
func GetTransfer(id string) (*Transfer, bool) {
	TransfersMutex.RLock()
	defer TransfersMutex.RUnlock()
	transfer, exists := ActiveTransfers[id]
	return transfer, exists
}


// RemoveTransfer removes a completed or failed transfer
func RemoveTransfer(id string) {
	TransfersMutex.Lock()
	defer TransfersMutex.Unlock()
	delete(ActiveTransfers, id)
}

// UpdateTransferStatus updates the status of a transfer
func UpdateTransferStatus(id string, status TransferStatus) {
	transfer, exists := GetTransfer(id)
	if !exists {
		return
	}
	
	transfer.PauseLock.Lock()
	defer transfer.PauseLock.Unlock()
	
	transfer.Status = status
}

type CheckpointedReader struct {
	Reader     io.Reader
	BytesRead  int64
	Transfer   *Transfer
	ChunkSize  int
	Buffer     []byte
	PauseCheck func() bool
}

// NewCheckpointedReader creates a new CheckpointedReader
func NewCheckpointedReader(reader io.Reader, transfer *Transfer, chunkSize int) *CheckpointedReader {
	return &CheckpointedReader{
		Reader:    reader,
		Transfer:  transfer,
		ChunkSize: chunkSize,
		Buffer:    make([]byte, chunkSize),
		PauseCheck: func() bool {
			transfer.PauseLock.Lock()
			defer transfer.PauseLock.Unlock()
			return transfer.IsPaused
		},
	}
}

// Read implements io.Reader and supports pausing
func (cr *CheckpointedReader) Read(p []byte) (n int, err error) {
	// Check if transfer is paused
	if cr.PauseCheck() {
		// Sleep a bit and check again to avoid CPU spinning
		time.Sleep(500 * time.Millisecond)
		return 0, nil
	}
	
	// Perform actual read
	n, err = cr.Reader.Read(p)
	
	if n > 0 {
		cr.BytesRead += int64(n)
		cr.Transfer.BytesComplete = cr.BytesRead
	}
	
	return n, err
}

// CheckpointedWriter is an io.Writer that supports pausing/resuming
type CheckpointedWriter struct {
	Writer      io.Writer
	BytesWritten int64
	Transfer    *Transfer
	ChunkSize   int
	Buffer      []byte
	PauseCheck  func() bool
}

// NewCheckpointedWriter creates a new CheckpointedWriter
func NewCheckpointedWriter(writer io.Writer, transfer *Transfer, chunkSize int) *CheckpointedWriter {
	return &CheckpointedWriter{
		Writer:     writer,
		Transfer:   transfer,
		ChunkSize:  chunkSize,
		Buffer:     make([]byte, chunkSize),
		PauseCheck: func() bool {
			transfer.PauseLock.Lock()
			defer transfer.PauseLock.Unlock()
			return transfer.IsPaused
		},
	}
}

// Write implements io.Writer and supports pausing
func (cw *CheckpointedWriter) Write(p []byte) (n int, err error) {
	if cw.PauseCheck() {
		time.Sleep(500 * time.Millisecond)
		return 0, nil
	}
	
	n, err = cw.Writer.Write(p)
	
	if n > 0 {
		cw.BytesWritten += int64(n)
		cw.Transfer.BytesComplete = cw.BytesWritten
	}
	
	return n, err
}
