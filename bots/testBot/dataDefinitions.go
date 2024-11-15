package main

type BotConfig struct {
	ApiKey    string `xml:"ApiKey"`
	SecretKey string `xml:"SecretKey"`
	BaseURL   string `xml:"BaseURL"`

	PairSymbol           string  `xml:"PairSymbol"`
	Symbol               string  `xml:"Symbol"`
	TradeAmount          float64 `xml:"TradeAmount"`
	ProfitPriceDelta     float64 `xml:"ProfitPriceDelta"`
	StopLossDelta        float64 `xml:"StopLossDelta"`
	MinWalletTradeAmount float64 `xml:"MinWalletTradeAmount"`

	FilePath string `xml:"FilePath"`
}
