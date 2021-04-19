package main

import (
	"github.com/gin-gonic/gin"
	"main.go/config"
	"main.go/controller"
)

func main() {
	config.LoadEnv()

	sess := config.ConnectAws()
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("sess", sess)
		c.Next()
	})

	router.POST("/upload", controller.UploadFile)
	router.POST("/uploads", controller.UploadFiles)
	router.DELETE("/delete", controller.DeleteFile)
	router.GET("/download",controller.DownloadFiles)

	_ = router.Run(":3030")
}
