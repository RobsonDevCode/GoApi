package models

import (
	"golang.org/x/net/websocket"
)

type User struct {
	ID       string          `json:"id"`
	Username string          `json:"user_name"`
	Conn     *websocket.Conn `json:"connection"` // WebSocket connection for real-time updates
}

// TickerDetailsRequest struct used to accept params for tickerDetails api call
type TickerDetailsRequest struct {
	Ticker string `form:"ticker" binding:"required"`
}

// GetFavouriteStocksOpenCloseRequest struct used to accept params for GetFavouriteStocksOpenClose api call
type GetFavouriteStocksOpenCloseRequest struct {
	UserId string `form:"user_id" binding:"required"` //we use string as uuid isn't safe for urls
}

type Subscription struct {
	UserId string `json:"user_id"`
	Ticker string `json:"ticker"`
}

type FavouriteStock struct {
	UserId string `form:"user_id" json:"user_id" binding:"required"`
	Ticker string `form:"ticker" json:"ticker" binding:"required"`
}
