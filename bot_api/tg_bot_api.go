package botapi

import (
	"CoinPriceNotifierBot/config"
	"fmt"
	"strconv"

	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

const (
	ON_COMMAND_NAME             = "/on"
	OFF_COMMAND_NAME            = "/off"
	SET_COIN_PRICE_COMMAND_NAME = "/set_coin_price"
)

type BotApi struct {
	api         *tgbotapi.BotAPI
	usersConfig map[int64]*config.UserConfig
}

func NewBotApi(accessToken string) (*BotApi, error) {
	api, err := tgbotapi.NewBotAPI(accessToken)
	if err != nil {
		return nil, err
	}

	return &BotApi{api: api, usersConfig: make(map[int64]*config.UserConfig)}, nil
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

			switch update.Message.Text {
			case ON_COMMAND_NAME:
				go b.executeOnCommand(update.Message)
			case OFF_COMMAND_NAME:
				go b.executeOffCommand(update.Message)
			case SET_COIN_PRICE_COMMAND_NAME:
				go b.executeSetCoinPriceCommand(update.Message)
			default:
				go b.executeDefaultMsgText(update.Message)
			}
		default:
			time.Sleep(time.Millisecond * 500)
		}
	}
}

func (b *BotApi) runUserGoroutine(userChatId int64) {
	for {
		select {
		case <-b.usersConfig[userChatId].GoroutineCh:
			return
		default:
			if b.usersConfig[userChatId].ChoosenCommand == "" {
				b.api.Send(tgbotapi.NewMessage(userChatId, fmt.Sprintf("Your coin price is: %f", b.usersConfig[userChatId].CoinPrice)))
			}
			time.Sleep(time.Second * time.Duration(b.usersConfig[userChatId].Timeout))
		}
	}
}

func (b *BotApi) executeDefaultMsgText(msg *tgbotapi.Message) {
	userConfig, exists := b.usersConfig[msg.Chat.ID]

	if exists {
		switch userConfig.ChoosenCommand {
		case SET_COIN_PRICE_COMMAND_NAME:
			b.executeSetCoinPriceValue(msg)
			return
		}
	}
	b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "Unknown command: "+strings.ReplaceAll(msg.Text, "\n", " ")))
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
	userConfig, exists := b.usersConfig[msg.Chat.ID]
	if exists && userConfig.HasActiveSession {
		stopCh := userConfig.GoroutineCh
		stopCh <- struct{}{}

		userConfig.HasActiveSession = false
		b.usersConfig[msg.Chat.ID] = userConfig

		b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "Successfully stop session"))
		return
	}
	b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "No active session"))
}

func (b *BotApi) executeSetCoinPriceCommand(msg *tgbotapi.Message) {
	userConfig, exists := b.usersConfig[msg.Chat.ID]
	if exists && userConfig.HasActiveSession {
		userConfig.ChoosenCommand = SET_COIN_PRICE_COMMAND_NAME

		b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "Send float price value"))
		return
	}
	b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "No active session"))
}

func (b *BotApi) executeSetCoinPriceValue(msg *tgbotapi.Message) {
	userConfig, exists := b.usersConfig[msg.Chat.ID]
	if exists && userConfig.HasActiveSession {
		floatPrice, err := strconv.ParseFloat(msg.Text, 64)
		if err != nil {
			b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Cannot parse %s to float", msg.Text)))
			userConfig.ChoosenCommand = ""
			return
		}
		userConfig.CoinPrice = floatPrice
		userConfig.ChoosenCommand = ""

		stopCh := userConfig.GoroutineCh
		stopCh <- struct{}{}

		go b.runUserGoroutine(msg.Chat.ID)
		b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Successfully changed coin price to: %f", floatPrice)))
		return
	}
	b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "No active session"))
}
