package models

import (
	"golang.org/x/net/websocket"
)

type User struct {
	ID       string          `json:"id"`
	Username string          `json:"username"`
	Conn     *websocket.Conn `json:"connection"` // WebSocket connection for real-time updates
}

// TickerDetailsRequest struct used to accept params for tickerDetails api call
type TickerDetailsRequest struct {
	Ticker string `form:"ticker" binding:"required"`
}

type Subscription struct {
	UserId string `json:"user_id"`
	Ticker string `json:"ticker"`
}
