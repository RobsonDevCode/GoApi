package stock

import (
	"fmt"
	. "github.com/RobsonDevCode/GoApi/cmd/api/dtos"
	"github.com/RobsonDevCode/GoApi/cmd/api/internal/processing"
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

var cache = &sync.Map{}

// GetTickerDetails gets stock information by ticker
func GetTickerDetails(c *gin.Context, pa *intergration.PolygonApi) {

	ctx := c.Request.Context()
	var request TickerDetailsDto

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
		cache.Load(key); ok {
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

		//cache timeout
		go func() {
			time.Sleep(4 * time.Minute)
			cache.Delete(key)
		}()

		//store result in cache
		cache.Store(key, result.Data)

		c.JSON(http.StatusOK, gin.H{"data": result.Data})

	case <-ctx.Done():
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request cancelled or timed out!"})
		return
	}

}

func GetSimpleMovingAverageForFavourites(c *gin.Context, stockDb StockRepository, pa *intergration.PolygonApi) {
}

func GetSimpleMovingAverage(c *gin.Context, pa *intergration.PolygonApi) {
	ctx := c.Request.Context()
	var params SimpleMovingAverageDto

	if err := c.ShouldBindQuery(&params); err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"validation error": err})
	}

	var wg sync.WaitGroup
	movingAvgCh := make(chan *Response[*polyModels.GetSMAResponse], 1)
	lwTickerCh := make(chan *Response[*polyModels.GetDailyOpenCloseAggResponse], 1)

	go func() {
		defer wg.Done()
		movingAverage := pa.FetchSimpleMovingAverage(params, ctx)

		movingAvgCh <- &movingAverage
	}()
	go func() {
		defer wg.Done()
		//go get the price of the tickers close last week
		weekAgo := params.TimeStamp.AddDate(0, 0, -7)
		lwTickerPrice := pa.FetchTickerOpenClose(params.Ticker, weekAgo, ctx)

		lwTickerCh <- &lwTickerPrice
	}()

	//wait for both to complete
	wg.Wait()

	//check if either channel errored
	movingAverage := <-movingAvgCh
	if movingAverage.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": movingAverage.Error.Error()})
		return
	}

	lwTickerPrice := <-lwTickerCh
	if lwTickerPrice.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": lwTickerPrice.Error.Error()})
		return
	}

	//api returns a list no matter the count, we're requesting one moving average across the week so
	//we grab the first and only value from the slice
	if params.Window <= 5 {
		avg := movingAverage.Data.Results.Values[0]
		//calculate the change between last week's close and this week's average price
		percentageDiff, err := processing.ComputePriceDelta(lwTickerPrice.Data.Close, avg.Value)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		result := &MovingAverageDto{
			SimpleMovingAverage:      *movingAverage.Data,
			PercentageChangeThisWeek: percentageDiff,
			InvestIndicator:          "Bullish",
		}

		c.JSON(http.StatusOK, gin.H{"success": result})
	}
}

// GetPreviousDayClose gets the previous day close for a ticker
func GetPreviousDayClose(c *gin.Context, pa *intergration.PolygonApi) {
	ctx := c.Request.Context()
	var params PreviousCloseRequestDto

	if err := c.ShouldBindQuery(&params); err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"validation error": err})
	}

	respCh := make(chan *Response[*polyModels.GetPreviousCloseAggResponse], 1)

	go func() {
		previousClose := pa.FetchPreviousClose(params, ctx)

		respCh <- &previousClose
	}()

	select {
	case result := <-respCh:
		if result.Error != nil {
			log.Error(result.Error)
			c.JSON(http.StatusBadRequest, gin.H{"error": result.Error.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": result.Data})
		return

	case <-ctx.Done():
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request cancelled or timed out!"})
		return
	}
}

// GetFavouriteStocksOpenClose gets favourite stocks open and close prices concurrently
func GetFavouriteStocksOpenClose(c *gin.Context, stockDb StockRepository, pa *intergration.PolygonApi) {
	ctx := c.Request.Context()
	var params GetFavouriteStocksOpenCloseDto

	if err := c.ShouldBindQuery(&params); err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"validation error": "user_id cannot be null when requesting favourite stocks"})
		return
	}

	//check if result has been cached
	key := fmt.Sprintf("get-fav-open-close-%s", params.UserId)

	if cacheResult, ok := cache.Load(key); ok {
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
			cache.Delete(key)
		}()
		//cache result
		cache.Store(key, response)
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
