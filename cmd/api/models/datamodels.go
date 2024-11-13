package models

import "golang.org/x/net/websocket"

type User struct {
	ID       string          `json:"id"`
	Username string          `json:"username"`
	Conn     *websocket.Conn `json:"connection"` // WebSocket connection for real-time updates
}

type Stock struct {
	Ticker       string           `json:"ticker"`
	Subscribers  map[string]*User `json:"subscribers"`
	PriceChannel chan float64     `json:"price_channel"` // Channel for broadcasting price updates
	StopChannel  chan bool        `json:"stop_channel"`  // Channel to stop the goroutine
}

type Subscription struct {
	UserId string `json:"user_id"`
	Ticker string `json:"ticker"`
}
