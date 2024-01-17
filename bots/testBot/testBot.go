package main

import (
	"context"
	"encoding/xml"
	"log"
	"math"
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

	var lastPrice float64 = 0

	for {
		walletAmmount, err := GetWalletAmount("USDT")

		if err != nil {
			ErrorLogger.Println(err.Error())
			return
		}
		if walletAmmount > 0 {
			cPrice, err := LastPrice(botConfig.PairSymbol)
			if err != nil {
				ErrorLogger.Println(err.Error())
				return
			}
			if (math.Abs(lastPrice-cPrice) >= math.Abs(lastPrice*botConfig.ProfitPriceDelta-lastPrice)) || (lastPrice == 0) {
				if cPrice*botConfig.TradeAmount < walletAmmount {
					NewOrderPair(botConfig.PairSymbol, botConfig.TradeAmount, botConfig.ProfitPriceDelta)
					lastPrice = cPrice
				}
			}
		}
		fileInfo, _ := os.Stat(botConfig.FilePath)
		cSavedTime := fileInfo.ModTime()
		if cSavedTime != lastSaveTime {
			botConfig = readConfig(os.Args[2])
			lastSaveTime = cSavedTime
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
		ErrorLogger.Println(err.Error())
		return
	}
	InfoLogger.Println("New Buy Order:", newOrder)
	//fmt.Println(binance_connector.PrettyPrint(newOrder))
	cLastPrice, err := LastPrice(pairSymbol)
	if err != nil {
		ErrorLogger.Println(err.Error())
		return
	}

	sellPrice := round(cLastPrice*priceDelta, 2)

	newSellOrder, err := client.NewCreateOrderService().Symbol(pairSymbol).
		Side("SELL").Type("LIMIT").Quantity(quantity).Price(sellPrice).TimeInForce("GTC").
		Do(context.Background())

	if err != nil {
		ErrorLogger.Println(err.Error())
		return
	}
	//fmt.Println(binance_connector.PrettyPrint(newSellOrder))
	InfoLogger.Println("New Sell Order:", newSellOrder)
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
