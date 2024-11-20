package stock

import (
	"fmt"
	. "github.com/RobsonDevCode/GoApi/cmd/api/internal/repository/dataAccess"
	stockConcurrency "github.com/RobsonDevCode/GoApi/cmd/api/internal/services/stock/concurrency"
	. "github.com/RobsonDevCode/GoApi/cmd/api/models"
	intergration "github.com/RobsonDevCode/GoApi/cmd/api/polygonApi"
	"github.com/gin-gonic/gin"
	"github.com/labstack/gommon/log"
	polyModels "github.com/polygon-io/client-go/rest/models"
	"net/http"
	"sync"
	"time"
)

const maxParallelism = 10 //this is set to 5 as that is the max amount of calls we get with the free version

var tickerDetailCache = &sync.Map{}
var tickerOpenCloseCache = &sync.Map{}

// GetTickerDetails gets stock information by ticker
func GetTickerDetails(c *gin.Context, pa *intergration.PolygonApi) {

	ctx := c.Request.Context()
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

	respChan := make(chan *Response[*polyModels.GetTickerDetailsResponse], 1)

	go func() {
		stocksDetails := pa.FetchTickerDetails(request.Ticker, ctx)

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

// GetFavouriteStocksOpenClose gets favourite stocks open and close prices concurrently
func GetFavouriteStocksOpenClose(c *gin.Context, stockDb StockRepository, pa *intergration.PolygonApi) {
	ctx := c.Request.Context()
	var params GetFavouriteStocksOpenCloseRequest

	if err := c.ShouldBindQuery(&params); err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"validation error": "user_id cannot be null when requesting favourite stocks"})
		return
	}

	//check if result has been cached
	key := fmt.Sprintf("get-fav-open-close-%s", params.UserId)

	if cacheResult, ok := tickerOpenCloseCache.Load(key); ok {
		c.JSON(http.StatusOK, cacheResult)
		return
	}

	respChan := make(chan *Response[[]string], 1)

	//launch go routine to go get tickers
	go func() {
		favouriteStocks := stockDb.GetFavouriteTickers(params.UserId, ctx)

		respChan <- &favouriteStocks
	}()

	select {
	case favouriteStocks := <-respChan:
		if favouriteStocks.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": favouriteStocks.Error.Error()})
			return
		}

		processor := stockConcurrency.NewPolyDataProcessor(pa, 10)

		resultCh, err := processor.ProcessTickersConcurrently(ctx, favouriteStocks.Data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var response []*polyModels.GetDailyOpenCloseAggResponse
		var errs []error

		for resp := range resultCh {
			if resp.Error != nil {
				errs = append(errs, resp.Error)
				log.Error(resp.Error)
				continue
			}

			response = append(response, resp.Data)
		}

		if len(errs) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": errs})
		}
		if len(response) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid responses received"})
			return
		}

		//set timer for cache
		go func() {
			time.Sleep(3 * time.Minute)
			tickerOpenCloseCache.Delete(key)
		}()
		//cache result
		tickerOpenCloseCache.Store(key, response)
		c.JSON(http.StatusOK, gin.H{"data": response})

	case <-ctx.Done():
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request cancelled or timed out!"})
		return

	}
}

// FavouriteTicker sets stock as a favourite for the user
func FavouriteTicker(c *gin.Context, stockDb StockRepository) {
	ctx := c.Request.Context()
	var request FavouriteStock

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error(err)
		if request.UserId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id cannot be null when creating a favourite stock"})
			return
		} else if request.Ticker == "" {
			c.JSON(http.StatusOK, gin.H{"error": "must provide a ticker"})
			return
		}

		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	ch := make(chan error)
	go func() {
		err := stockDb.AddToFavouriteTickers(request, ctx)
		ch <- err
	}()

	select {
	case err := <-ch:
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	case <-ctx.Done():
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request cancelled or timed out!"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"created": "Stock added to favorites"})
	return
}

// UnFavouriteTicker removes a stock from favourites
func UnFavouriteTicker(c *gin.Context, stockDb StockRepository) {
	ctx := c.Request.Context()
	var request FavouriteStock

	if err := c.ShouldBindQuery(&request); err != nil {
		log.Error(err)
		if request.UserId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id cannot be null when removing a stock from favourites"})
			return
		} else if request.Ticker == "" {
			c.JSON(http.StatusOK, gin.H{"error": "must provide a ticker"})
			return
		}
	}

	ch := make(chan error)
	go func() {
		err := stockDb.DeleteStockFromFavouriteTickers(request, ctx)
		ch <- err
	}()

	select {
	case err := <-ch:
		if err != nil {
			log.Errorf("error removing favourite: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	case <-ctx.Done():
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request cancelled or timed out!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": "Stock removed from favorites"})
	return
}
