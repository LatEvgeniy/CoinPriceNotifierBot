package config

type UserConfig struct {
	Timeout            int
	NotificationFormat string
	CoinPrice          float64
	GoroutineCh        chan struct{}
	ChoosenCommand     string
	HasActiveSession   bool
}

func GetDefaultConfig() *UserConfig {
	return &UserConfig{Timeout: 1, NotificationFormat: "2006-01-02 15:04:05"}
}
