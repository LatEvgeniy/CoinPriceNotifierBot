package api

import (
	"CoinPriceNotifierBot/config"
	"CoinPriceNotifierBot/dto"
	"fmt"
	"os"
	"strconv"

	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	ON_COMMAND_NAME                        = "/on"
	OFF_COMMAND_NAME                       = "/off"
	SET_COIN_PRICE_COMMAND_NAME            = "/set_coin_price"
	SET_COIN_PRICE_SCALE_COMMAND_NAME      = "/set_coin_price_scale"
	SET_NOTIFICATION_INTERVAL_COMMAND_NAME = "/set_notification_interval"

	MAX_USER_PRICE_SCALE = 16
	MIN_USER_PRICE_SCALE = 0

	MAX_USER_SECONDS_INTERVAL = 86_400
	MIN_USER_SECONDS_INTERVAL = 2

	EXECUTE_TG_UPDATES_MILLIS_TIMEOUT = 500

	API_CURRENCY_NAME_BITCOIN = "bitcoin"
	NO_ACTIVE_SESSION_ERR_MSG = "No active session"

	BITCOIN_PRICES_FILE_NAME = "prices_history/Prices_Bitcoin.txt"
	TIME_FORMAT              = time.DateTime
)

type BotApi struct {
	api               *tgbotapi.BotAPI
	usersConfig       map[int64]*config.UserConfig
	cryptoCurrencyApi *CryptoCurrencyApi
	sessionDataApi    *SessionDataApi
}

func NewBotApi(accessToken string, crCryptoCurrencyApi *CryptoCurrencyApi, sessionDataApi *SessionDataApi) (*BotApi, error) {
	api, err := tgbotapi.NewBotAPI(accessToken)
	if err != nil {
		return nil, err
	}

	return &BotApi{api: api, usersConfig: make(map[int64]*config.UserConfig), cryptoCurrencyApi: crCryptoCurrencyApi, sessionDataApi: sessionDataApi}, nil
}

func (b *BotApi) Run() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates, err := b.api.GetUpdatesChan(updateConfig)
	if err != nil {
		logrus.Errorln(err)
		panic(err)
	}

	for {
		select {
		case update := <-updates:
			logrus.Infof(
				"Received message: %s from %s with chat id: %d",
				strings.ReplaceAll(update.Message.Text, "\n", " "),
				update.Message.From.UserName,
				update.Message.Chat.ID,
			)

			if usrConf, exists := b.usersConfig[update.Message.Chat.ID]; exists && usrConf.ChoosenCommand != "" {
				go b.executeAnswerMsgText(update.Message)
				continue
			}

			switch update.Message.Text {
			case ON_COMMAND_NAME:
				go b.executeOnCommand(update.Message)
			case OFF_COMMAND_NAME:
				go b.executeOffCommand(update.Message)
			case SET_COIN_PRICE_COMMAND_NAME:
				go b.executeSetCoinPriceCommand(update.Message)
			case SET_COIN_PRICE_SCALE_COMMAND_NAME:
				go b.executeSetCoinPriceScaleCommand(update.Message)
			case SET_NOTIFICATION_INTERVAL_COMMAND_NAME:
				go b.executeSetNotificationIntervalCommand(update.Message)
			}
		default:
			time.Sleep(time.Millisecond * EXECUTE_TG_UPDATES_MILLIS_TIMEOUT)
		}
	}
}

func (b *BotApi) runUserGoroutine(userChatId int64) {
	ticker := time.NewTicker(time.Second * time.Duration(b.usersConfig[userChatId].Timeout))
	defer ticker.Stop()

	for {
		select {
		case <-b.usersConfig[userChatId].GoroutineCh:
			return
		case <-ticker.C:
			if b.usersConfig[userChatId].ChoosenCommand == "" {
				b.getCoinPrice(userChatId)
			}
		}
	}
}

func (b *BotApi) executeAnswerMsgText(msg *tgbotapi.Message) {
	if userConfig, exists := b.usersConfig[msg.Chat.ID]; exists {
		switch userConfig.ChoosenCommand {
		case SET_COIN_PRICE_COMMAND_NAME:
			b.executeSetCoinPriceValue(msg)
		case SET_COIN_PRICE_SCALE_COMMAND_NAME:
			b.executeSetCoinPriceScaleValue(msg)
		case SET_NOTIFICATION_INTERVAL_COMMAND_NAME:
			b.executeSetNotificationIntervalValue(msg)
		default:
			b.executeDefault(msg) // call get session data
		}
	}
}

func (b *BotApi) getCoinPrice(userChatId int64) {
	englishTitle := cases.Title(language.English)

	bitcoinPirce, err := b.cryptoCurrencyApi.getCurrencyPrice(API_CURRENCY_NAME_BITCOIN, b.usersConfig[userChatId].CoinPriceScale)
	if err != nil {
		b.api.Send(tgbotapi.NewMessage(userChatId, fmt.Sprintf("Error while getting currency price: %s", err.Error())))
	}

	floatUserCointPrice := b.usersConfig[userChatId].CoinPrice

	stringUserCoinPrice := strconv.FormatFloat(floatUserCointPrice, 'f', -1, 64)
	if floatUserCointPrice == float64(int64(floatUserCointPrice)) {
		stringUserCoinPrice = fmt.Sprintf("%.0f", floatUserCointPrice)
	}

	msgToUser := fmt.Sprintf(
		"Saved %s price = %s\n%s price now = %s",
		API_CURRENCY_NAME_BITCOIN,
		stringUserCoinPrice,
		englishTitle.String(API_CURRENCY_NAME_BITCOIN),
		bitcoinPirce,
	)

	b.api.Send(tgbotapi.NewMessage(userChatId, msgToUser))

	b.writePriceToFile(BITCOIN_PRICES_FILE_NAME, bitcoinPirce)
}

func (b *BotApi) executeOnCommand(msg *tgbotapi.Message) {
	userConfig, exists := b.usersConfig[msg.Chat.ID]
	if exists && userConfig.HasActiveSession {
		b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "Session already exists, to run new session - stop the previous one"))
		return
	}

	if !exists {
		config := config.GetDefaultConfig()
		config.GoroutineCh = make(chan struct{})

		b.usersConfig[msg.Chat.ID] = config
	}
	b.usersConfig[msg.Chat.ID].HasActiveSession = true

	b.runUserGoroutine(msg.Chat.ID)
}

func (b *BotApi) executeOffCommand(msg *tgbotapi.Message) {
	if userConfig, exists := b.usersConfig[msg.Chat.ID]; exists && userConfig.HasActiveSession {
		b.usersConfig[msg.Chat.ID].GoroutineCh <- struct{}{}
		b.usersConfig[msg.Chat.ID].HasActiveSession = false

		b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "Successfully stop session"))
		return
	}
	b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, NO_ACTIVE_SESSION_ERR_MSG))
}

func (b *BotApi) executeSetCoinPriceCommand(msg *tgbotapi.Message) {
	if userConfig, exists := b.usersConfig[msg.Chat.ID]; exists && userConfig.HasActiveSession {
		b.usersConfig[msg.Chat.ID].ChoosenCommand = SET_COIN_PRICE_COMMAND_NAME

		b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "Send float price value"))
		return
	}
	b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, NO_ACTIVE_SESSION_ERR_MSG))
}

func (b *BotApi) executeSetCoinPriceValue(msg *tgbotapi.Message) {
	if userConfig, exists := b.usersConfig[msg.Chat.ID]; exists && userConfig.HasActiveSession {
		floatPrice, err := strconv.ParseFloat(msg.Text, 64)
		if err != nil {
			b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Cannot parse %s to float", msg.Text)))
			b.usersConfig[msg.Chat.ID].ChoosenCommand = ""
			return
		}
		b.usersConfig[msg.Chat.ID].CoinPrice = floatPrice
		b.usersConfig[msg.Chat.ID].ChoosenCommand = ""

		b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Successfully changed coin price to: %f", floatPrice)))
		return
	}
	b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, NO_ACTIVE_SESSION_ERR_MSG))
}

func (b *BotApi) executeSetCoinPriceScaleCommand(msg *tgbotapi.Message) {
	if userConfig, exists := b.usersConfig[msg.Chat.ID]; exists && userConfig.HasActiveSession {
		b.usersConfig[msg.Chat.ID].ChoosenCommand = SET_COIN_PRICE_SCALE_COMMAND_NAME

		userMsg := fmt.Sprintf("Send coin price scale value that must be %d <= scale <= %d", MIN_USER_PRICE_SCALE, MAX_USER_PRICE_SCALE)

		b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, userMsg))
		return
	}
	b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, NO_ACTIVE_SESSION_ERR_MSG))
}

func (b *BotApi) executeSetCoinPriceScaleValue(msg *tgbotapi.Message) {
	scale, err := strconv.Atoi(msg.Text)
	b.usersConfig[msg.Chat.ID].ChoosenCommand = ""

	if err != nil {
		b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Cannot parse %s to int", msg.Text)))
		return
	}

	if scale > MAX_USER_PRICE_SCALE || scale < MIN_USER_PRICE_SCALE {
		b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Scale must be %d <= scale <= %d", MIN_USER_PRICE_SCALE, MAX_USER_PRICE_SCALE)))
		return
	}

	b.usersConfig[msg.Chat.ID].CoinPriceScale = scale
	b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Successfully changed coin price  scale to: %d", scale)))
}

func (b *BotApi) executeSetNotificationIntervalCommand(msg *tgbotapi.Message) {
	if userConfig, exists := b.usersConfig[msg.Chat.ID]; exists && userConfig.HasActiveSession {
		b.usersConfig[msg.Chat.ID].ChoosenCommand = SET_NOTIFICATION_INTERVAL_COMMAND_NAME

		userMsg := fmt.Sprintf("Send seconds interval that must be %d <= interval <= %d", MIN_USER_SECONDS_INTERVAL, MAX_USER_SECONDS_INTERVAL)

		b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, userMsg))
		return
	}
	b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, NO_ACTIVE_SESSION_ERR_MSG))
}

func (b *BotApi) executeSetNotificationIntervalValue(msg *tgbotapi.Message) {
	interval, err := strconv.Atoi(msg.Text)
	b.usersConfig[msg.Chat.ID].ChoosenCommand = ""

	if err != nil {
		b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Cannot parse %s to int", msg.Text)))
		return
	}

	if interval > MAX_USER_SECONDS_INTERVAL || interval < MIN_USER_SECONDS_INTERVAL {
		b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Interval must be %d <= interval <= %d", MIN_USER_SECONDS_INTERVAL, MAX_USER_SECONDS_INTERVAL)))
		return
	}

	b.usersConfig[msg.Chat.ID].Timeout = interval
	b.usersConfig[msg.Chat.ID].GoroutineCh <- struct{}{}

	go b.runUserGoroutine(msg.Chat.ID)
	b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Successfully changed interval to: %d", interval)))
}

func (b *BotApi) writePriceToFile(fileName, price string) error {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		file, err := os.Create(fileName)
		if err != nil {
			return err
		}
		file.Close()
	}

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	currentTime := time.Now().Format(TIME_FORMAT)

	if _, err := file.WriteString(fmt.Sprintf("%s %d %s\n", currentTime, time.Now().Unix(), price)); err != nil {
		return err
	}

	return nil
}

func (b *BotApi) executeDefault(msg *tgbotapi.Message) { // call get session data
	request := &dto.SessionDataRequestDto{NamePerson: strings.Split(msg.From.UserName, "@")[1], ToDo: msg.Text}

	sessionData, err := b.sessionDataApi.getSessionData(request)
	if err != nil {
		b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, err.Error()))
		return
	}

	b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, sessionData))
}
