package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/shellhaki/opensecret/modules/receiver"
	"github.com/shellhaki/opensecret/modules/sender"
)

func main() {
	r := gin.Default()

	r.POST("/transfer/receive", receiver.HandleReceiver)
	r.GET("/transfer/receive/:id", receiver.HandleTransferStatus)
	r.GET("/transfer/receive/:id/download", receiver.HandleDownload)
	r.GET("/transfer/send/:id", sender.HandleSenderInfo)
	r.POST("/transfer/send/:id", sender.HandleSender)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
