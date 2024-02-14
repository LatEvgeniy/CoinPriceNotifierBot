package config

type UserConfig struct {
	Timeout            int
	NotificationFormat string
	UserCoinPrice      int
}

func GetDefaultConfig() *UserConfig {
	return &UserConfig{Timeout: 1, NotificationFormat: "2006-01-02 15:04:05", UserCoinPrice: 0}
}
