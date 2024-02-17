package config

type UserConfig struct {
	Timeout          int // 0 - 86_400
	CoinPrice        float64
	CoinPriceScale   int // 0 - 16
	GoroutineCh      chan struct{}
	ChoosenCommand   string
	HasActiveSession bool
}

func GetDefaultConfig() *UserConfig {
	return &UserConfig{Timeout: 3, CoinPriceScale: 2}
}
