package routing

import (
	repository "github.com/RobsonDevCode/GoApi/cmd/api/internal/repository/dataAccess"
	"github.com/RobsonDevCode/GoApi/cmd/api/internal/routing/stocks"
	integration "github.com/RobsonDevCode/GoApi/cmd/api/polygonApi"
	"github.com/gin-gonic/gin"
	"github.com/labstack/gommon/log"
)

func NewRouter(stockRepo *repository.StockRepository) error {
	log.Info("api starting up...")
	router := gin.Default()

	log.Info("connecting to polygon api...")
	polyClient := integration.ConnectToPolygonApi()

	//Stock Controller
	stockHandler := stocks.SetUpStockHandler(stockRepo, polyClient)
	stockHandler.RegisterRoutes(router)

	server := "localhost:8080"
	err := router.Run(server)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Infof("api sucessfully running on %s", server)

	return nil
}
