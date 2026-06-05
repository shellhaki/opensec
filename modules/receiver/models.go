package receiver

import "time"

type TransferStatus string

const (
	TransferWaiting   TransferStatus = "waiting"
	TransferReceiving TransferStatus = "receiving"
	TransferCompleted TransferStatus = "completed"
)

type FileTransfer struct {
	ID          string         `json:"id"`
	Filename    string         `json:"filename,omitempty"`
	Size        int64          `json:"size,omitempty"`
	Path        string         `json:"-"`
	Status      TransferStatus `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
}
