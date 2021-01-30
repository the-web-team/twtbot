package stocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"os"
)

type Service struct {
	Session *discordgo.Session
}

type PolygonResponse struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
	Ticker    struct {
		Day struct {
			OpenPrice              float64 `json:"o"`
			HighestPrice           float64 `json:"h"`
			LowestPrice            float64 `json:"l"`
			ClosePrice             float64 `json:"c"`
			TradingVolume          float64 `json:"v"`
			VolumeWeightedAvgPrice float64 `json:"vw"`
		} `json:"day"`
		LastTrade struct {
			TradeConditions []int   `json:"c"`
			TradeID         string  `json:"i"`
			Price           float64 `json:"p"`
			Volume          float64 `json:"s"`
			StartTimestamp  int     `json:"t"`
			ExchangeID      int     `json:"x"`
		} `json:"lastTrade"`
		Min struct {
			AccumulatedVolume      int     `json:"av"`
			OpenPrice              float64 `json:"o"`
			HighestPrice           float64 `json:"h"`
			LowestPrice            float64 `json:"l"`
			ClosePrice             float64 `json:"c"`
			TradingVolume          float64 `json:"v"`
			VolumeWeightedAvgPrice float64 `json:"vw"`
		} `json:"min"`
		PrevDay struct {
			OpenPrice              float64 `json:"o"`
			HighestPrice           float64 `json:"h"`
			LowestPrice            float64 `json:"l"`
			ClosePrice             float64 `json:"c"`
			TradingVolume          float64 `json:"v"`
			VolumeWeightedAvgPrice float64 `json:"vw"`
		} `json:"prevDay"`
		Ticker             string  `json:"ticker"`
		TodayChange        float64 `json:"todaysChange"`
		TodayChangePercent float64 `json:"todaysChangePerc"`
		LastUpdated        int     `json:"updated"`
	} `json:"ticker"`
}

func (r *Service) CheckStockPrice(symbol string) (*PolygonResponse, error) {
	polygonApiKey := os.Getenv("POLYGON_API_KEY")
	response, _ := http.Get(fmt.Sprintf("https://api.polygon.io/v2/snapshot/locale/global/markets/crypto/tickers/X:%sUSD?&apiKey=%s", symbol, polygonApiKey))
	buf := new(bytes.Buffer)
	_, readError := buf.ReadFrom(response.Body)
	if readError != nil {
		return nil, readError
	}
	str := buf.String()
	fmt.Println(str)
	var parsed PolygonResponse
	jsonError := json.Unmarshal(buf.Bytes(), &parsed)
	if jsonError != nil {
		return nil, jsonError
	}

	return &parsed, nil
}
