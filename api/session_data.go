package api

import (
	dto "CoinPriceNotifierBot/dto"
	"encoding/json"

	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"
)

type SessionDataApi struct {
	client *resty.Client
	url    string
}

func NewSessionDataApi(url string) *SessionDataApi {
	return &SessionDataApi{client: resty.New(), url: url}
}

func (c *SessionDataApi) getSessionData(request *dto.SessionDataRequestDto) (string, error) {
	resp, err := c.client.R().SetBody(request).Post(c.url)
	if err != nil {
		logrus.Errorf("Ошибка при отправке запроса: %v", err)
		return "", err
	}

	if resp.StatusCode() != 200 {
		logrus.Errorf("Ошибка при выполнении запроса. Код ответа: %d", resp.StatusCode())
		return "", err
	}

	var sessionData dto.SessionDataResponseDto
	if err := json.Unmarshal(resp.Body(), &sessionData); err != nil {
		logrus.Fatal(err)
		return "", err
	}

	return sessionData.DataText, nil
}
