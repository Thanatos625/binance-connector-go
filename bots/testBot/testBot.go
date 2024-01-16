package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"time"

	binance_connector "github.com/binance/binance-connector-go"
)

var apiKey = ""
var secretKey = ""
var baseURL = ""

func main() {
	if len(os.Args) > 1 && os.Args[2] != "" {
		botConfig := readConfig(os.Args[2])
		apiKey = botConfig.ApiKey
		secretKey = botConfig.SecretKey
		baseURL = botConfig.BaseURL

		StartBot(botConfig.PairSymbol, botConfig.Symbol, botConfig.TradeAmount, botConfig.ProfitPriceDelta)
	} else {
		fmt.Println("Missing config argument")
	}
}

func StartBot(pairSymbol, symbol string, tradeAmount, profitPriceDelta float64) {

	var lastPrice float64 = 0

	for {
		walletAmmount, _ := GetWalletAmount("USDT")

		if walletAmmount > 0 {
			cPrice, _ := LastPrice(pairSymbol)
			if (math.Abs(lastPrice-cPrice) >= math.Abs(lastPrice*profitPriceDelta)) || (lastPrice == 0) {
				if cPrice*tradeAmount < walletAmmount {
					NewOrderPair(pairSymbol, tradeAmount, profitPriceDelta)
					lastPrice = cPrice
				}
			}
		}
		time.Sleep(15 * time.Second)
	}
}

func NewOrderPair(pairSymbol string, quantity, priceDelta float64) {

	client := binance_connector.NewClient(apiKey, secretKey, baseURL)
	// Create new order
	newOrder, err := client.NewCreateOrderService().Symbol(pairSymbol).
		Side("BUY").Type("MARKET").Quantity(quantity).
		Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(binance_connector.PrettyPrint(newOrder))
	cLastPrice, err := LastPrice(pairSymbol)
	if err != nil {
		fmt.Println(err)
		return
	}

	sellPrice := round(cLastPrice*(1+priceDelta), 2)

	newSellOrder, err := client.NewCreateOrderService().Symbol(pairSymbol).
		Side("SELL").Type("LIMIT").Quantity(quantity).Price(sellPrice).TimeInForce("GTC").
		Do(context.Background())

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(binance_connector.PrettyPrint(newSellOrder))
}

func readConfig(filepath string) BotConfig {
	var config BotConfig
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("Error opening XML file: %v\n", err)
		return config
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		fmt.Printf("Error decoding XML: %v\n", err)
		return config
	}
	return config
}
