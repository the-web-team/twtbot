package getstockprice

import (
	"fmt"
	"twtbot/interfaces"
	"twtbot/services/stocks"
)

type Handler struct {
	interfaces.MessageHandler
}

func (h *Handler) ShouldRun() bool {
	return h.CommandHasPrefix('$')
}

func (h *Handler) Run() error {
	symbol := h.GetPrefixedCommand('$')
	s := stocks.Service{Session: h.Session}
	price, err := s.CheckStockPrice(symbol)
	if err != nil {
		return err
	}
	reply := fmt.Sprintf("%s is currently at `$%f`!", symbol, price)
	return h.ReplyWithMention(reply)
}
