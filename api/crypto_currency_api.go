package api

import (
	model "CoinPriceNotifierBot/model"
	"encoding/json"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

type CryptoCurrencyApi struct {
	client *resty.Client
	url    string
}

func NewCryptoCurrencyApi(url string) *CryptoCurrencyApi {
	return &CryptoCurrencyApi{client: resty.New(), url: url}
}

func (c *CryptoCurrencyApi) getCurrencyPrice(currencyName string, currencyPriceScale int) (string, error) {
	resp, err := c.client.R().Get(c.url + currencyName)
	if err != nil {
		logrus.Fatalf("Ошибка при отправке запроса: %v", err)
		return "", err
	}

	if resp.StatusCode() != 200 {
		logrus.Fatalf("Ошибка при выполнении запроса. Код ответа: %d", resp.StatusCode())
		return "", err
	}

	var coinCupModel model.CoinCupModel
	if err := json.Unmarshal([]byte(resp.Body()), &coinCupModel); err != nil {
		logrus.Fatal(err)
		return "", err
	}

	decimalPrice, err := decimal.NewFromString(coinCupModel.Data.PriceUsd)
	if err != nil {
		logrus.Fatal(err)
		return "", err
	}
	roundedPrice := decimalPrice.Round(int32(currencyPriceScale))

	return roundedPrice.String(), nil
}
