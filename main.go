package main

import (
	"CoinPriceNotifierBot/api"
	botapi "CoinPriceNotifierBot/api"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	GET_PRICE_API_URL        = "https://api.coincap.io/v2/assets/"
	FILE_NAME_WITH_BOT_TOKEN = "bot_token.txt"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	cryptoCurrencyApi := api.NewCryptoCurrencyApi(GET_PRICE_API_URL)

	botToken, err := getBotTokenFromFile(FILE_NAME_WITH_BOT_TOKEN)
	if err != nil {
		logrus.Errorln(err)
		panic(err)
	}

	botApi, err := botapi.NewBotApi(botToken, cryptoCurrencyApi)
	if err != nil {
		logrus.Errorln(err)
		panic(err)
	}

	logrus.Debugln("Succesfully started bot")
	botApi.Run()
}

func getBotTokenFromFile(fileName string) (string, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		logrus.Errorln(err)
		panic(err)
	}

	return string(data), nil
}
