package services

import (
	"context"
	"fmt"
	. "github.com/RobsonDevCode/GoApi/cmd/api/models"
	. "github.com/RobsonDevCode/GoApi/cmd/api/settings/configuration"
	"github.com/gin-gonic/gin"
	"github.com/labstack/gommon/log"
	polygon "github.com/polygon-io/client-go/rest"
	polyModels "github.com/polygon-io/client-go/rest/models"
	"net/http"
	"sync"
	"time"
)

var tickerDetailCache = &sync.Map{}

// GetTickerDetails gets stock information by ticker
func GetTickerDetails(c *gin.Context) {

	var request TickerDetailsRequest

	if err := c.ShouldBindQuery(&request); err != nil {
		//handle bad request
		log.Error(err)
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err})
		return
	}
	//check if result has been cached
	key := fmt.Sprintf("ticker-details-%s", request.Ticker)

	if cacheResult, ok :=
		tickerDetailCache.Load(key); ok {
		//data has been cached already
		c.JSON(http.StatusOK, cacheResult)
		return
	}

	respChan := make(chan *ApiResponse[*polyModels.GetTickerDetailsResponse], 1)
	ctx := c.Request.Context()

	go func() {
		stocksDetails := fetchTickerDetails(request.Ticker, ctx)

		respChan <- &stocksDetails
	}()

	select {
	case result := <-respChan:
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
			return
		}

		//tickerDetailCache timeout
		go func() {
			time.Sleep(4 * time.Minute)
			tickerDetailCache.Delete(key)
		}()

		//store result in tickerDetailCache
		tickerDetailCache.Store(key, result.Data)

		c.JSON(http.StatusOK, gin.H{"data": result.Data})

	case <-ctx.Done():
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request cancelled or timed out!"})
		return
	}

}

func FavouriteTicker(c *gin.Context) {

}
func fetchTickerDetails(ticker string, ctx context.Context) ApiResponse[*polyModels.GetTickerDetailsResponse] {
	c := polygon.New(Configuration.ApiSettings.Key)

	params := &polyModels.GetTickerDetailsParams{
		Ticker: ticker,
	}

	//make request
	response, err := c.GetTickerDetails(ctx, params)
	if err != nil {
		log.Errorf("Error calling GetLastTrade: %s", err)

		result := ApiResponse[*polyModels.GetTickerDetailsResponse]{
			Data:  nil,
			Error: err,
		}
		return result
	}

	log.Info(response)

	result := &ApiResponse[*polyModels.GetTickerDetailsResponse]{
		Data:  response,
		Error: nil,
	}

	return *result
}
