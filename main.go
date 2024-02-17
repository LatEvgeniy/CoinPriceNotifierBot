package main

import (
	"CoinPriceNotifierBot/api"
	botapi "CoinPriceNotifierBot/api"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	GET_PRICE_API_URL = "https://api.coincap.io/v2/assets/"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	cryptoCurrencyApi := api.NewCryptoCurrencyApi(GET_PRICE_API_URL)

	botApi, err := botapi.NewBotApi(os.Args[1], cryptoCurrencyApi)
	if err != nil {
		logrus.Errorln(err)
		panic(err)
	}

	logrus.Debugln("Succesfully started bot")
	botApi.Run()
}
