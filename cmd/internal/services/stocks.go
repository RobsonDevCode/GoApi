package services

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetStocks(c *gin.Context) {
	respChan := make(chan *string)

	go func() {
		resp := "working chan"

		respChan <- &resp
	}()

	resp := <-respChan

	c.IndentedJSON(http.StatusOK, resp)
}
