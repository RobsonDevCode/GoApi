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
		stockHandler.GET("tickerDetails", func(c *gin.Context) {
			stock.GetTickerDetails(c, s.polyClient)
		})
		stockHandler.GET("/getFavouriteStocksOpenClose", func(c *gin.Context) {
			stock.GetFavouriteStocksOpenClose(c, *s.stockRepo, s.polyClient)
		})

		//********** POST/PUT/PATCH COMMANDS **********
		stockHandler.POST("/addToFavourites", func(c *gin.Context) {
			stock.FavouriteTicker(c, *s.stockRepo)
		})

		//********** DELETE COMMANDS**********
		stockHandler.DELETE("/removeFromFavourites", func(c *gin.Context) {
			stock.UnFavouriteTicker(c, *s.stockRepo)
		})

	}

}
