package sender

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shellhaki/opensecret/modules/receiver"
)

func HandleSenderInfo(c *gin.Context) {
	id := c.Param("id")

	if !receiver.HasTransfer(id) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Transfer not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transfer_id":  id,
		"method":       "POST",
		"field":        "file",
		"send_url":     fmt.Sprintf("/transfer/send/%s", id),
		"download_url": fmt.Sprintf("/transfer/receive/%s/download", id),
	})
}

func HandleSender(c *gin.Context) {
	id := c.Param("id")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Upload a file using multipart form field 'file'",
		})
		return
	}

	transfer, err := receiver.StoreFile(id, file)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to store uploaded file"

		switch {
		case errors.Is(err, receiver.ErrTransferNotFound):
			status = http.StatusNotFound
			message = "Transfer not found"
		case errors.Is(err, receiver.ErrTransferUnavailable):
			status = http.StatusConflict
			message = "Transfer already has a file or is currently receiving one"
		case errors.Is(err, receiver.ErrFileMissing):
			status = http.StatusBadRequest
			message = "Uploaded file is missing"
		}

		c.JSON(status, gin.H{
			"error": message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "File uploaded successfully",
		"transfer":     transfer,
		"download_url": fmt.Sprintf("/transfer/receive/%s/download", id),
	})
}
