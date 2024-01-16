package main

import (
	"context"
	"fmt"
	"math"
	"strconv"

	binance_connector "github.com/binance/binance-connector-go"
)

func GetWalletAmount(symbol string) (float64, error) {
	client := binance_connector.NewClient(apiKey, secretKey, baseURL)
	asset, err := client.NewUserAssetService().Asset(symbol).
		Do(context.Background())

	if err != nil {
		fmt.Println(err)
		return -1, err
	}
	fmt.Println(binance_connector.PrettyPrint(asset))

	retVal, err := strconv.ParseFloat(asset[len(asset)-1].Free, 64)
	if err != nil {
		fmt.Println("Error:", err)
		return retVal, err
	}

	return retVal, nil
}
func LastPrice(symbol string) (float64, error) {

	client := binance_connector.NewClient("", "", baseURL)

	// AvgPrice
	lastPrice, err := client.NewTickerService().
		Symbol(symbol).Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	fmt.Println(binance_connector.PrettyPrint(lastPrice))

	fLastPrice, err := strconv.ParseFloat(lastPrice.LastPrice, 64)
	if err != nil {
		fmt.Println("Error:", err)
		return 0, err
	}

	return fLastPrice, nil
}

// round rounds a float64 to a specified number of decimal places
func round(f float64, precision int) float64 {
	shift := math.Pow(10, float64(precision))
	return math.Round(f*shift) / shift
}
func GetCurrentOpenOrders() {

	client := binance_connector.NewClient(apiKey, secretKey, baseURL)

	// Binance Get current open orders - GET /api/v3/openOrders
	getCurrentOpenOrders, err := client.NewGetOpenOrdersService().Symbol("BTCUSDT").
		Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(binance_connector.PrettyPrint(getCurrentOpenOrders))
}
func NewBuyOrder() {

	client := binance_connector.NewClient(apiKey, secretKey, baseURL)
	// Create new order
	newOrder, err := client.NewCreateOrderService().Symbol("BTCUSDT").
		Side("BUY").Type("MARKET").Quantity(0.00012).
		Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(binance_connector.PrettyPrint(newOrder))
}
func NewSellOrder() {

	client := binance_connector.NewClient(apiKey, secretKey, baseURL)
	// Create new order
	newOrder, err := client.NewCreateOrderService().Symbol("BTCUSDT").
		Side("SELL").Type("MARKET").Quantity(0.00012).
		Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(binance_connector.PrettyPrint(newOrder))
}
func CancelOrder() {

	client := binance_connector.NewClient(apiKey, secretKey, baseURL)

	// Binance Cancel Order endpoint - DELETE /api/v3/order
	cancelOrder, err := client.NewCancelOrderService().Symbol("BTCUSDT").OrderId(24304541478).
		Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(binance_connector.PrettyPrint(cancelOrder))
}
