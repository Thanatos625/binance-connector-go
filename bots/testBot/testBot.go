package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"time"

	binance_connector "github.com/binance/binance-connector-go"
)

var apiKey = ""
var secretKey = ""
var baseURL = ""
var lastSaveTime time.Time
var botConfig BotConfig

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

func main() {
	if len(os.Args) > 1 && os.Args[2] != "" {
		initLogger()
		botConfig = readConfig(os.Args[2])
		apiKey = botConfig.ApiKey
		secretKey = botConfig.SecretKey
		baseURL = botConfig.BaseURL

		StartBot()

	} else {
		ErrorLogger.Println("Missing config argument")
	}
}

func StartBot() {
	var lastWalletAmmout float64 = 0

	for {
		var quantity = botConfig.TradeAmount
		walletAmmount, err := GetWalletAmount("USDT")

		if err != nil {
			ErrorLogger.Println(err.Error())
			return
		}
		if walletAmmount > botConfig.MinWalletTradeAmount {
			if cPrice, constInc, quantity, hasError := getMaxBuying(walletAmmount, lastWalletAmmout, quantity); !hasError {
				if cPrice*quantity < walletAmmount && quantity > constInc {
					NewOrderPair(botConfig.PairSymbol, quantity, botConfig.ProfitPriceDelta, botConfig.StopLossDelta)
				}
			}
		}
		reloadConfig()

		time.Sleep(15 * time.Second)
	}
}

func reloadConfig() {
	fileInfo, _ := os.Stat(botConfig.FilePath)
	cSavedTime := fileInfo.ModTime()
	if cSavedTime != lastSaveTime {
		botConfig = readConfig(os.Args[2])
		lastSaveTime = cSavedTime
	}
}

func getMaxBuying(walletAmmount float64, lastWalletAmmout float64, quantity float64) (float64, float64, float64, bool) {
	cPrice, err := LastPrice(botConfig.PairSymbol)
	if err != nil {
		ErrorLogger.Println(err.Error())
		return 0, 0, 0, true
	}
	if walletAmmount != lastWalletAmmout {
		InfoLogger.Println("Wallet Amount:", walletAmmount)
		lastWalletAmmout = walletAmmount
	}
	constInc := 0.00001
	if quantity < 0 {

		for inc := 0.0; cPrice*inc <= walletAmmount; inc = inc + constInc {
			quantity = roundWithDecimals(inc, 6)
		}
		quantity -= constInc
		quantity = roundWithDecimals(quantity, 6)
	} else {
		if cPrice*botConfig.TradeAmount < walletAmmount {
			quantity = botConfig.TradeAmount
		} else {
			quantity = -1
		}
	}
	return cPrice, constInc, quantity, false
}

func NewOrderPair(pairSymbol string, quantity, priceDelta float64, stopLossDelta float64) bool {

	client := binance_connector.NewClient(apiKey, secretKey, baseURL)

	getCurrentOpenOrders, err := client.NewGetOpenOrdersService().Symbol(pairSymbol).
		Do(context.Background())
	if err != nil {
		ErrorLogger.Println(err)
		return false
	}
	if len(getCurrentOpenOrders) > 0 {
		cancelOpenOrders, err := client.NewCancelOpenOrdersService().Symbol(pairSymbol).
			Do(context.Background())
		if err != nil {
			ErrorLogger.Println(err)
		}
		InfoLogger.Println(binance_connector.PrettyPrint(cancelOpenOrders))
		return false
	}

	newOrder, err := client.NewCreateOrderService().Symbol(pairSymbol).
		Side("BUY").Type("MARKET").Quantity(quantity).
		Do(context.Background())
	if err != nil {
		ErrorLogger.Println(err.Error())
		return false
	}
	s, _ := json.Marshal(newOrder)
	InfoLogger.Println("New Buy Order:", string(s))
	fmt.Println(binance_connector.PrettyPrint(newOrder))

	var walletAmmountBTC float64

	if waitForOrderToFullfill(client, pairSymbol, quantity) {
		return false
	}

	cLastPrice, err := LastPrice(pairSymbol)
	if err != nil {
		ErrorLogger.Println(err.Error())
		return false
	}
	walletAmmountBTC, err = GetWalletAmount("BTC")
	if err != nil {
		ErrorLogger.Println(err.Error())
		return false
	}
	walletAmmountBTC = floorWithDecimals(walletAmmountBTC, 5)

	sellPriceUP := roundWithDecimals(cLastPrice*priceDelta, 0)
	sellPriceDown := roundWithDecimals(cLastPrice/stopLossDelta, 0)
	sellStopPriceDown := roundWithDecimals(cLastPrice/stopLossDelta, 0)
	sellStopPriceDown = roundWithDecimals(sellStopPriceDown*(1-0.002), 0) //with 0.2 under stop loose price

	// newOCO, err := client.NewNewOCOService().Symbol("BTCUSDT").
	// 	Side("SELL").Quantity(0.002).Price(69650).StopPrice(69600).StopLimitPrice(69550).StopLimitTimeInForce("GTC").
	// 	Do(context.Background())
	newOCO, err := client.NewNewOCOService().Symbol(pairSymbol).Side("SELL").StopLimitTimeInForce("GTC").
		Quantity(walletAmmountBTC).Price(sellPriceUP).StopPrice(sellPriceDown).StopLimitPrice(sellStopPriceDown).
		Do(context.Background())
	if err != nil {
		ErrorLogger.Println(err.Error())
		return false
	}
	s, _ = json.Marshal(newOCO)
	InfoLogger.Println("New OCO Order:", string(s))

	// newSellOrder, err := client.NewCreateOrderService().Symbol(pairSymbol).
	// 	Side("SELL").Type("LIMIT").Quantity(walletAmmountBTC).Price(sellPriceUP).TimeInForce("GTC").
	// 	Do(context.Background())
	// if err != nil {
	// 	ErrorLogger.Println(err.Error())
	// 	return false
	// }
	// s, _ = json.Marshal(newSellOrder)
	// InfoLogger.Println("New Sell Take Profit Order:", string(s))

	return true
}

func waitForOrderToFullfill(client *binance_connector.Client, pairSymbol string, quantity float64) bool {
	for {
		getCurrentOpenOrders, err := client.NewGetOpenOrdersService().Symbol(pairSymbol).
			Do(context.Background())
		if err != nil {
			ErrorLogger.Println(err)
			return true
		}
		walletAmmountBTC, err := GetWalletAmount("BTC")

		if err != nil {
			ErrorLogger.Println(err.Error())
			return true
		}

		if len(getCurrentOpenOrders) == 0 && walletAmmountBTC >= quantity {
			break
		}
		time.Sleep(time.Second * 1)
	}
	return false
}

func readConfig(filepath string) BotConfig {
	var config BotConfig
	file, err := os.Open(filepath)
	if err != nil {
		ErrorLogger.Println(err.Error())
		return config
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		ErrorLogger.Println(err.Error())
		return config
	}
	fileInfo, _ := os.Stat(filepath)
	lastSaveTime = fileInfo.ModTime()
	config.FilePath = filepath

	InfoLogger.Println("Reload Configuration")
	return config
}
func initLogger() {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
