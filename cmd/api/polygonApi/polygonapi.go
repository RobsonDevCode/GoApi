package integration

import (
	"context"
	"errors"
	"github.com/RobsonDevCode/GoApi/cmd/api/dtos"
	. "github.com/RobsonDevCode/GoApi/cmd/api/models"
	. "github.com/RobsonDevCode/GoApi/cmd/api/settings/configuration"
	"time"

	"github.com/labstack/gommon/log"
	polygon "github.com/polygon-io/client-go/rest"
	polyModels "github.com/polygon-io/client-go/rest/models"
)

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
		log.Errorf("Error calling ticker details: %s", err)
		return Response[*polyModels.GetTickerDetailsResponse]{
			Data:  nil,
			Error: err,
		}
	}

}

func (p *PolygonApi) FetchPreviousClose(dto dtos.PreviousCloseRequestDto, ctx context.Context) Response[*polyModels.GetPreviousCloseAggResponse] {
	params := &polyModels.GetPreviousCloseAggParams{
		Ticker:   dto.Ticker,
		Adjusted: &dto.Adjusted,
	}
	if response, err := p.client.GetPreviousCloseAgg(ctx, params); err == nil {
		return Response[*polyModels.GetPreviousCloseAggResponse]{
			Data:  response,
			Error: nil,
		}
	} else {
		log.Errorf("error calling previous close: %s", err)
		return Response[*polyModels.GetPreviousCloseAggResponse]{
			Data:  nil,
			Error: err,
		}
	}

}

func (p *PolygonApi) FetchTickerOpenClose(ticker string, dateFrom time.Time, ctx context.Context) Response[*polyModels.GetDailyOpenCloseAggResponse] {

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	params := &polyModels.GetDailyOpenCloseAggParams{
		Ticker: ticker,
		Date:   polyModels.Date(dateFrom),
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

func (p *PolygonApi) FetchSimpleMovingAverage(request dtos.SimpleMovingAverageDto, ctx context.Context) Response[*polyModels.GetSMAResponse] {

	params := &polyModels.GetSMAParams{
		Ticker:           request.Ticker,
		TimestampGTE:     (*polyModels.Millis)(&request.TimeStamp),
		Timespan:         (*polyModels.Timespan)(&request.TimeSpan),
		Window:           &request.Window,
		ExpandUnderlying: &request.MoreDetails,
	}

	if response, err := p.client.GetSMA(ctx, params); err == nil {
		result := Response[*polyModels.GetSMAResponse]{
			Data:  response,
			Error: nil,
		}
		if len(result.Data.Results.Values) == 0 {
			result.Error = errors.New("result from simple moving average was empty")
			return result
		}

		return result

	} else {
		log.Errorf("Error calling SMA: %s", err)
		return Response[*polyModels.GetSMAResponse]{
			Data:  nil,
			Error: err,
		}
	}
}
