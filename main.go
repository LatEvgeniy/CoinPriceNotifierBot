package main

import (
	botapi "CoinPriceNotifierBot/api"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	CONFIG_FILE_NAME             = "config/config.yaml"
	BOT_TOKEN_CONFIG_KEY         = "bot-token"
	SESSION_DATA_URL_CONFIG_KEY  = "session-data-url"
	GET_PRICE_API_URL_CONFIG_KEY = "get-price-api-url"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	viper.SetConfigFile(CONFIG_FILE_NAME)

	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatalf("Error while reading config file: %s by path %s", err, CONFIG_FILE_NAME)
		panic(err)
	}

	sessionDataApi := botapi.NewSessionDataApi(viper.GetString(SESSION_DATA_URL_CONFIG_KEY))
	cryptoCurrencyApi := botapi.NewCryptoCurrencyApi(viper.GetString(GET_PRICE_API_URL_CONFIG_KEY))

	botApi, err := botapi.NewBotApi(viper.GetString(BOT_TOKEN_CONFIG_KEY), cryptoCurrencyApi, sessionDataApi)
	if err != nil {
		logrus.Errorln(err)
		panic(err)
	}

	logrus.Debugln("Succesfully started bot")
	botApi.Run()
}
