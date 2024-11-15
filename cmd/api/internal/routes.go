package routes

import (
	"github.com/RobsonDevCode/GoApi/cmd/api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/labstack/gommon/log"
)

func NewRouter() error {
	log.Info("api starting up...")
	router := gin.Default()

	router.GET("/stocks:ticker", services.GetStocks)
	server := "localhost:8080"
	err := router.Run(server)
	if err != nil {
		log.Fatal(err)
		return err
	}

	log.Infof("api sucessfully running on %s", server)

	return nil
}
