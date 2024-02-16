package main

import (
	botapi "CoinPriceNotifierBot/api"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	botApi, err := botapi.NewBotApi(os.Args[1])
	if err != nil {
		logrus.Errorln(err)
		panic(err)
	}

	logrus.Debugln("Succesfully started bot")
	botApi.Run()
}
