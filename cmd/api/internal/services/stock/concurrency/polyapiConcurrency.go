package stockConcurrency

import (
	"context"
	"github.com/RobsonDevCode/GoApi/cmd/api/models"
	intergration "github.com/RobsonDevCode/GoApi/cmd/api/polygonApi"
	"github.com/labstack/gommon/log"
	polyModels "github.com/polygon-io/client-go/rest/models"
	"sync"
)

// PolyDataProcessor handles concurrent processing of stock data from or to the polygon api
type PolyDataProcessor struct {
	api            *intergration.PolygonApi
	maxParallelism int8
}

// NewPolyDataProcessor creates a new StockDataProcessor
func NewPolyDataProcessor(api *intergration.PolygonApi, maxParallelism int8) *PolyDataProcessor {
	return &PolyDataProcessor{
		api:            api,
		maxParallelism: maxParallelism,
	}
}

// ProcessTickersConcurrently processes multiple tickers concurrently with rate limiting
func (p *PolyDataProcessor) ProcessTickersConcurrently(ctx context.Context, tickers []string) (<-chan models.Response[*polyModels.GetDailyOpenCloseAggResponse], error) {
	resultCh := make(chan models.Response[*polyModels.GetDailyOpenCloseAggResponse], p.maxParallelism)
	var wg sync.WaitGroup

	p.spanWorkerPool(ctx, &wg, resultCh, tickers)

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	return resultCh, nil
}

// spanWorkerPool creates the worker pool with rate limiting
func (p *PolyDataProcessor) spanWorkerPool(ctx context.Context, wg *sync.WaitGroup,
	resultCh chan<- models.Response[*polyModels.GetDailyOpenCloseAggResponse], tickers []string) {
	semaphore := make(chan struct{}, p.maxParallelism)

	for _, ticker := range tickers {
		wg.Add(1)
		t := ticker // Capture for closure

		go p.processTickerWorker(ctx, t, semaphore, resultCh, wg)
	}

}

// processTickerWorker handles individual ticker processing
func (p *PolyDataProcessor) processTickerWorker(ctx context.Context, ticker string, semaphore chan struct{},
	resultCh chan<- models.Response[*polyModels.GetDailyOpenCloseAggResponse], wg *sync.WaitGroup) {

	defer wg.Done()
	semaphore <- struct{}{}
	defer func() { <-semaphore }()

	log.Infof("Processing ticker %s", ticker)

	response := p.api.FetchTickerOpenClose(ticker)

	select {
	case resultCh <- response:
		log.Debugf("Sent response for ticker: %s", ticker)
	case <-ctx.Done():
		log.Errorf("Context cancelled while sending response for ticker %s", ticker)
	}
}
