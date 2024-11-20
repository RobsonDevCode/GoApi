package repository

import (
	"context"
	"database/sql"
	"fmt"
	. "github.com/RobsonDevCode/GoApi/cmd/api/models"
	"github.com/labstack/gommon/log"
)

type StocksDataBase struct {
	*sql.DB
}

type StockRepository struct {
	db *StocksDataBase
}

func NewStockRepository(db *StocksDataBase) *StockRepository {
	return &StockRepository{db: db}
}

func (s *StockRepository) AddToFavouriteTickers(favouriteStock FavouriteStock, ctx context.Context) error {

	query := "INSERT INTO favourite_tickers (id, ticker) VALUES (?, ?)"

	result, err := s.db.ExecContext(ctx, query, favouriteStock.UserId, favouriteStock.Ticker)
	if err != nil {
		log.Errorf("error executing query: %s", err)
		return err
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		log.Errorf("error checking rows affected: %s", err)
		return err
	}
	if rowsAff != 1 {
		return fmt.Errorf("query executed, but change to db does not match. Rows Affected: %d", rowsAff)

	}

	return nil
}
func (s *StockRepository) DeleteStockFromFavouriteTickers(favouriteStock FavouriteStock, ctx context.Context) error {
	query := "DELETE FROM favourite_tickers WHERE id = ? AND  ticker = ?"

	rows, err := s.db.ExecContext(ctx, query, favouriteStock.UserId, favouriteStock.Ticker)
	if err != nil {
		log.Errorf("error executing remove from favourite query: %s", err)
		return err
	}

	//check if the query was actually successful
	row, err := rows.RowsAffected()
	if err != nil {
		log.Errorf("error checking rows affected: %s", err)
		return err
	}
	if row != 1 {
		rowErr := fmt.Errorf("query executed, but change to db does not match. Rows Affected: %d", row)
		return rowErr
	}

	log.Infof("%s, was removed from %s favourites", favouriteStock.Ticker, favouriteStock.UserId)
	return nil
}
func (s *StockRepository) GetFavouriteTickers(id string, ctx context.Context) Response[[]string] {
	query := "SELECT ticker FROM favourite_tickers WHERE id = ?"

	rows, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		log.Errorf("error executing query: %s", err)

		return Response[[]string]{
			Data:  nil,
			Error: err,
		}
	}
	defer rows.Close()

	var tickers []string
	var ticker string

	for rows.Next() {
		if err := rows.Scan(&ticker); err != nil {
			log.Errorf("error scanning row: %s", err)
			return Response[[]string]{
				Data:  nil,
				Error: err,
			}
		}
		tickers = append(tickers, ticker)
	}

	if len(tickers) == 0 {
		return Response[[]string]{
			Data:  nil,
			Error: fmt.Errorf("no favourite tickers found"),
		}
	}

	return Response[[]string]{
		Data:  tickers,
		Error: nil,
	}
}
