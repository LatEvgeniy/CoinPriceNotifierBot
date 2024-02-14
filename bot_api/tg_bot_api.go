package botapi

import (
	"CoinPriceNotifierBot/config"

	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

type BotApi struct {
	api           *tgbotapi.BotAPI
	userGoroutine map[int64]chan struct{}
	mutex         sync.Mutex
}

func NewBotApi(accessToken string) (*BotApi, error) {
	api, err := tgbotapi.NewBotAPI(accessToken)
	if err != nil {
		return nil, err
	}

	return &BotApi{api: api, userGoroutine: make(map[int64]chan struct{}), mutex: sync.Mutex{}}, nil
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
			logrus.Infof("Receive new message from %d: %s", update.Message.Chat.ID, update.Message.Text)
			switch update.Message.Text {
			case "/on":
				go b.executeOnCommand(update.Message)
			case "/off":
				go b.executeOffCommand(update.Message)
			default:
				go b.api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command: "+update.Message.Text))
			}
		default:
			time.Sleep(time.Millisecond * 500)
		}
	}
}

func (b *BotApi) executeOnCommand(msg *tgbotapi.Message) {
	stopCh := make(chan struct{})

	// add check unique id
	b.mutex.Lock()
	b.userGoroutine[msg.Chat.ID] = stopCh
	b.mutex.Unlock()

	defaultConfig := config.GetDefaultConfig()

	for {
		select {
		case <-stopCh:
			logrus.Debugln("Got stop msg")
			return
		default:
			logrus.Debugln(b.userGoroutine)
			b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "Executing..."))
			time.Sleep(time.Second * time.Duration(defaultConfig.Timeout))
		}
	}
}

func (b *BotApi) executeOffCommand(msg *tgbotapi.Message) {
	stopCh, ok := b.userGoroutine[msg.Chat.ID]
	if ok {
		logrus.Debugln("Sending stop msg")
		close(stopCh)
		delete(b.userGoroutine, msg.Chat.ID)
		logrus.Debugln("Sent stop msg")
		return
	}
	b.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "No active session"))
}
