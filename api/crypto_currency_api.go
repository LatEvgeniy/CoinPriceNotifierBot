package api

import (
	model "CoinPriceNotifierBot/model"
	"encoding/json"
	"log"

	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

type CryptoCurrencyApi struct {
	url    string
	client *resty.Client
}

func GetNewCryptoCurrencyApi(url string) *CryptoCurrencyApi {
	return &CryptoCurrencyApi{url: url}
}

func (c *CryptoCurrencyApi) getCurrencyPrice(currencyName string) string {
	resp, err := c.client.R().Get(c.url + currencyName)
	if err != nil {
		logrus.Fatalf("Ошибка при отправке запроса: %v", err)
		return ""
	}

	if resp.StatusCode() != 200 {
		logrus.Fatalf("Ошибка при выполнении запроса. Код ответа: %d", resp.StatusCode())
		return ""
	}

	// Инициализация структуры для маппинга данных
	var coinCupModel model.CoinCupModel

	// Распаковка JSON-данных в структуру
	if err := json.Unmarshal([]byte(resp.Body()), &coinCupModel); err != nil {
		log.Fatal(err)
		return ""
	}

	return coinCupModel.Data.PriceUsd
}
