package config

import "github.com/spf13/viper"

type Config struct {
	RabbitMQURL    string `mapstructure:"RABBITMQ_URL"`
	OneSignalKey   string `mapstructure:"ONESIGNAL_KEY"`
	OneSignalAppID string `mapstructure:"ONESIGNAL_APP_ID"`
	PostgresUrl    string `mapstructure:"POSTGRES_URL"`
	RedisURL       string `mapstructure:"REDIS_URL"`
	Port           string `mapstructure:"PORT"`
	ServiceName    string `mapstructure:"SERVICE_NAME"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	viper.AddConfigPath("../")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	var err error
	if err = viper.ReadInConfig(); err != nil {
		return nil, err
	}
	viper.Unmarshal(&cfg)
	return &cfg, nil
}
