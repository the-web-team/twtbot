package getstockprice

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
	"twtbot/interfaces"
	"twtbot/services/stocks"
)

type Handler struct {
	interfaces.MessageHandler
}

func (h *Handler) ShouldRun() bool {
	isStocksChannel := h.IsChannel("591677226001367051") || h.IsChannel("748976491630690436")
	return isStocksChannel && h.CommandHasPrefix('$')
}

func (h *Handler) Run() error {
	symbol := h.GetPrefixedCommand('$')
	fmt.Println(symbol)
	s := stocks.Service{Session: h.Session}
	result, err := s.CheckStockPrice(symbol)
	if err != nil {
		fmt.Println("Stocks error:", err)
		return err
	}
	return h.SendPrice(symbol, result)
}

func (h *Handler) SendPrice(symbol string, price *stocks.PolygonResponse) error {
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: "Cryptocurrency",
		},
		Title: fmt.Sprintf("<:dogecoin:804838402238185502> `%s:USD`", symbol),
		Description: strings.Join([]string{
			formatField("Last Trade Price", fmt.Sprintf("$%f", price.Ticker.LastTrade.Price)),
			formatField("Trading Volume", fmt.Sprintf("$%f", price.Ticker.Day.TradingVolume)),
			formatField("Previous Day Close Price", fmt.Sprintf("$%f", price.Ticker.PrevDay.ClosePrice)),
			formatField("Percent Change Today", fmt.Sprintf("%f%%", price.Ticker.TodayChangePercent)),
		}, "\n"),
		Timestamp: time.Now().Format("2006-01-02T15:04:05-0700"),
		Color:     0x009900,
		Provider: &discordgo.MessageEmbedProvider{
			URL:  "https://polygon.io",
			Name: "Polygon.io",
		},
	}
	_, messageError := h.Session.ChannelMessageSendEmbed(h.Message.ChannelID, embed)
	return messageError
}

func formatField(title string, value string) string {
	return fmt.Sprintf("**%s:** %s", title, value)
}
