package main

import (
	repository "github.com/RobsonDevCode/GoApi/cmd/api/internal/repository/dataAccess"
	"github.com/RobsonDevCode/GoApi/cmd/api/internal/routing"
	"github.com/RobsonDevCode/GoApi/cmd/api/settings/configuration"
	"log"
	"time"
)

func run() error {
	if err := configuration.SetEnvironmentSettings("development"); err != nil {
		log.Fatalf("Failed to set environment variables: %v", err)
		return err
	}

	//Add Connections and there configs
	stocksDB, err := configuration.NewDB(configuration.DbConfig{
		DSN:          configuration.Configuration.ConnectionStrings.Stocks,
		MaxOpenConns: 25,
		MaxIdelConns: 25,
		MaxLifeTime:  15 * time.Minute,
		MaxIdelTime:  5 * time.Minute,
	})

	if err != nil {
		log.Fatalf("error setting up databases: %s", err)
	}
	defer stocksDB.Close()

	stockRepo := repository.NewStockRepository(&repository.StocksDataBase{DB: stocksDB})

	routerErr := routing.NewRouter(stockRepo)
	if routerErr != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
