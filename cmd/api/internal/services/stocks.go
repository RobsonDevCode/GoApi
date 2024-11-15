package services

import (
	. "github.com/RobsonDevCode/GoApi/cmd/api/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
)

func GetStocks(c *gin.Context) {
	respChan := make(chan *ApiResponse[Stock])
	var wg sync.WaitGroup
	wg.Add(1)

	ticker := c.Param("ticker")

	go fetchStocks(ticker, respChan, &wg)

	go func() {
		wg.Wait()
		close(respChan)
	}()

	c.IndentedJSON(http.StatusOK, respChan)
}

func fetchStocks(ticker string, ch chan<- *ApiResponse[Stock], wg *sync.WaitGroup) *ApiResponse[Stock] {

	defer wg.Done()

	ch <- &ApiResponse[Stock]{}

	return &ApiResponse[Stock]{}
}
