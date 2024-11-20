package integration

import (
	"context"
	. "github.com/RobsonDevCode/GoApi/cmd/api/models"
	. "github.com/RobsonDevCode/GoApi/cmd/api/settings/configuration"
	"time"

	"github.com/labstack/gommon/log"
	polygon "github.com/polygon-io/client-go/rest"
	polyModels "github.com/polygon-io/client-go/rest/models"
)

var polyConfig = Settings{
	Yesterday: time.Now().AddDate(0, 0, -1),
}

type PolygonApi struct {
	client *polygon.Client
}

func ConnectToPolygonApi() *PolygonApi {
	return &PolygonApi{
		client: polygon.New(Configuration.ApiSettings.Key),
	}
}

func (p *PolygonApi) FetchTickerDetails(ticker string, ctx context.Context) Response[*polyModels.GetTickerDetailsResponse] {

	params := &polyModels.GetTickerDetailsParams{
		Ticker: ticker,
	}

	//make request to get ticker details https://polygon.io/docs/stocks/get_v3_reference_tickers__ticker
	if response, err := p.client.GetTickerDetails(ctx, params); err == nil {
		return Response[*polyModels.GetTickerDetailsResponse]{
			Data:  response,
			Error: nil,
		}
	} else {
		log.Errorf("Error calling Last Trade: %s", err)
		return Response[*polyModels.GetTickerDetailsResponse]{
			Data:  nil,
			Error: err,
		}
	}

}

func (p *PolygonApi) FetchTickerOpenClose(ticker string) Response[*polyModels.GetDailyOpenCloseAggResponse] {
	ctx := context.Background()
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	params := &polyModels.GetDailyOpenCloseAggParams{
		Ticker: ticker,
		Date:   polyModels.Date(polyConfig.Yesterday),
	}

	//make request to Daily open and close https://polygon.io/docs/stocks/get_v1_open-close__stocksticker___date
	if response, err := p.client.GetDailyOpenCloseAgg(ctx, params); err == nil {
		return Response[*polyModels.GetDailyOpenCloseAggResponse]{
			Data:  response,
			Error: nil,
		}
	} else {
		log.Errorf("Error calling Daily Open Close: %s", err)
		return Response[*polyModels.GetDailyOpenCloseAggResponse]{
			Data:  nil,
			Error: err,
		}
	}

}
