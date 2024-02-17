package config

type UserConfig struct {
	Timeout            int
	NotificationFormat string
	CoinPrice          float64
	CoinPriceScale     int // 0 - 16
	GoroutineCh        chan struct{}
	ChoosenCommand     string
	HasActiveSession   bool
}

func GetDefaultConfig() *UserConfig {
	return &UserConfig{Timeout: 3, NotificationFormat: "2006-01-02 15:04:05", CoinPriceScale: 2}
}
