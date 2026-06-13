package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func router() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.GET("/image", generateAndServeImage)
	router.Run(":8080")
}

func generateAndServeImage(c *gin.Context) {
	address := c.DefaultQuery("address", "none")
	if address == "none" {
		c.Status(http.StatusBadRequest)
		return
	}

	stats, err := getStatsByAddress(address, 12)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	serverInfo, ok := infoObjects[address]
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}

	buf, err := genImage(*serverInfo, stats)
	if err != nil {
		c.String(http.StatusInternalServerError, "ошибка генерации изображения: %v", err)
		return
	}
	c.Data(http.StatusOK, "image/png", buf.Bytes())
}
