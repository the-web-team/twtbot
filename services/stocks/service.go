package stocks

import (
	"github.com/bwmarrin/discordgo"
	"github.com/vic3lord/stocks"
)

type Service struct {
	Session *discordgo.Session
}

func (r *Service) CheckStockPrice(symbol string) (float64, error) {
	stock, getQuoteError := stocks.GetQuote(symbol)
	if getQuoteError != nil {
		return 0.0, getQuoteError
	}

	return stock.GetPrice()
}
