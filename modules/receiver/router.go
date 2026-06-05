package receiver

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const uploadDirectory = "received_files"

var (
	ErrFileMissing         = errors.New("file missing")
	ErrTransferNotFound    = errors.New("transfer not found")
	ErrTransferUnavailable = errors.New("transfer unavailable")
)

var (
	transfers   = make(map[string]*FileTransfer)
	transferMux sync.RWMutex
)

func HandleReceiver(c *gin.Context) {
	transfer, err := createTransfer()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create transfer session",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "File receive session created",
		"transfer":     transfer,
		"send_url":     fmt.Sprintf("/transfer/send/%s", transfer.ID),
		"status_url":   fmt.Sprintf("/transfer/receive/%s", transfer.ID),
		"download_url": fmt.Sprintf("/transfer/receive/%s/download", transfer.ID),
	})
}

func HandleTransferStatus(c *gin.Context) {
	transfer, exists := FindTransfer(c.Param("id"))
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Transfer not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transfer": transfer,
	})
}

func HandleDownload(c *gin.Context) {
	transfer, exists := FindTransfer(c.Param("id"))
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Transfer not found",
		})
		return
	}

	if transfer.Status != TransferCompleted || transfer.Path == "" {
		c.JSON(http.StatusConflict, gin.H{
			"error": "No file has been uploaded for this transfer yet",
		})
		return
	}

	if _, err := os.Stat(transfer.Path); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Uploaded file is no longer available",
		})
		return
	}

	c.FileAttachment(transfer.Path, transfer.Filename)
}

func HasTransfer(id string) bool {
	transferMux.RLock()
	defer transferMux.RUnlock()

	_, exists := transfers[id]
	return exists
}

func FindTransfer(id string) (*FileTransfer, bool) {
	transferMux.RLock()
	defer transferMux.RUnlock()

	transfer, exists := transfers[id]
	if !exists {
		return nil, false
	}

	snapshot := *transfer
	return &snapshot, true
}

func StoreFile(id string, header *multipart.FileHeader) (*FileTransfer, error) {
	if header == nil {
		return nil, ErrFileMissing
	}

	filename := filepath.Base(header.Filename)
	if filename == "." || filename == string(filepath.Separator) || filename == "" {
		return nil, ErrFileMissing
	}

	transferMux.Lock()
	transfer, exists := transfers[id]
	if !exists {
		transferMux.Unlock()
		return nil, ErrTransferNotFound
	}
	if transfer.Status != TransferWaiting {
		transferMux.Unlock()
		return nil, ErrTransferUnavailable
	}

	transfer.Status = TransferReceiving
	transfer.Filename = filename
	transfer.Size = header.Size
	transfer.Path = filepath.Join(uploadDirectory, id, filename)
	transferMux.Unlock()

	if err := saveUploadedFile(header, transfer.Path); err != nil {
		resetTransfer(id)
		return nil, err
	}

	completedAt := time.Now().UTC()

	transferMux.Lock()
	transfer.Status = TransferCompleted
	transfer.CompletedAt = &completedAt
	if stat, err := os.Stat(transfer.Path); err == nil {
		transfer.Size = stat.Size()
	}
	snapshot := *transfer
	transferMux.Unlock()

	return &snapshot, nil
}

func createTransfer() (*FileTransfer, error) {
	id, err := generateTransferID()
	if err != nil {
		return nil, err
	}

	transfer := &FileTransfer{
		ID:        id,
		Status:    TransferWaiting,
		CreatedAt: time.Now().UTC(),
	}

	transferMux.Lock()
	transfers[id] = transfer
	transferMux.Unlock()

	return transfer, nil
}

func generateTransferID() (string, error) {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func saveUploadedFile(header *multipart.FileHeader, destinationPath string) error {
	source, err := header.Open()
	if err != nil {
		return err
	}
	defer source.Close()

	if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
		return err
	}

	destination, err := os.Create(destinationPath)
	if err != nil {
		return err
	}

	if _, err := io.Copy(destination, source); err != nil {
		destination.Close()
		return err
	}

	return destination.Close()
}

func resetTransfer(id string) {
	transferMux.Lock()
	defer transferMux.Unlock()

	transfer, exists := transfers[id]
	if !exists {
		return
	}

	transfer.Status = TransferWaiting
	transfer.Filename = ""
	transfer.Size = 0
	transfer.Path = ""
	transfer.CompletedAt = nil
}
