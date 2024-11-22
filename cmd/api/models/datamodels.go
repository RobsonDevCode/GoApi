package models

type Subscription struct {
	UserId string `json:"user_id"`
	Ticker string `json:"ticker"`
}

type FavouriteStock struct {
	UserId string `form:"user_id" json:"user_id" binding:"required"`
	Ticker string `form:"ticker" json:"ticker" binding:"required"`
}
