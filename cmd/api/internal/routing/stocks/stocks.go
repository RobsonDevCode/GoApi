package stocks

import (
	repository "github.com/RobsonDevCode/GoApi/cmd/api/internal/repository/dataAccess"
	"github.com/RobsonDevCode/GoApi/cmd/api/internal/services/stock"
	intergration "github.com/RobsonDevCode/GoApi/cmd/api/polygonApi"
	"github.com/gin-gonic/gin"
)

type StockHandler struct {
	stockRepo  *repository.StockRepository
	polyClient *intergration.PolygonApi
}

func SetUpStockHandler(repo *repository.StockRepository, polyClient *intergration.PolygonApi) *StockHandler {
	return &StockHandler{
		stockRepo:  repo,
		polyClient: polyClient,
	}
}

func (s *StockHandler) RegisterRoutes(router *gin.Engine) {
	stockHandler := router.Group("/stocks")
	{
		//********** GET COMMANDS**********
		stockHandler.GET("info/tickerdetails", func(c *gin.Context) {
			stock.GetTickerDetails(c, s.polyClient)
		})
		stockHandler.GET("daily/openclose", func(c *gin.Context) {
			stock.GetFavouriteStocksOpenClose(c, *s.stockRepo, s.polyClient)
		})
		stockHandler.GET("daily/changeFromYesterday", func(c *gin.Context) {
			stock.GetPreviousDayClose(c, s.polyClient)
		})
		stockHandler.GET("indicators/sma", func(c *gin.Context) {
			stock.GetSimpleMovingAverage(c, s.polyClient)
		})

		//********** POST/PUT/PATCH COMMANDS **********
		stockHandler.POST("/favourites/add", func(c *gin.Context) {
			stock.FavouriteTicker(c, *s.stockRepo)
		})

		//********** DELETE COMMANDS**********
		stockHandler.DELETE("/favourites/delete", func(c *gin.Context) {
			stock.UnFavouriteTicker(c, *s.stockRepo)
		})

	}

}
