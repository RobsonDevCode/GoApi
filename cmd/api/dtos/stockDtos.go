package dtos

import (
	polyModels "github.com/polygon-io/client-go/rest/models"
	"time"
)

// TickerDetailsDto struct used to accept params for tickerDetails api call
type TickerDetailsDto struct {
	Ticker string `form:"ticker" binding:"required"`
}

// GetFavouriteStocksOpenCloseDto struct used to accept params for GetFavouriteStocksOpenClose api call
type GetFavouriteStocksOpenCloseDto struct {
	UserId string `form:"user_id" binding:"required"` //we use string as uuid isn't safe for urls
}

type PreviousCloseRequestDto struct {
	Ticker   string `form:"ticker" binding:"required"`
	Adjusted bool   `form:"adjusted" default:"true"`
}

type MovingAverageDto struct {
	SimpleMovingAverage      polyModels.GetSMAResponse `json:"simple_moving_average"`
	PercentageChangeThisWeek string                    `json:"percentage_change_this_week"`
	InvestIndicator          string                    `json:"invest_indicator"`
}

type SimpleMovingAverageDto struct {
	Ticker      string    `form:"ticker" binding:"required"`
	TimeStamp   time.Time `form:"time_stamp" binding:"required"`
	TimeSpan    string    `form:"time_span" binding:"required"`
	Window      int       `form:"window" binding:"required"`
	MoreDetails bool      `form:"more_details"  default:"false"`
}
