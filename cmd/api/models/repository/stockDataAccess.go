package models

import "github.com/google/uuid"

type FavouritedStocks struct {
	UserId  uuid.UUID
	Tickers []string
}
