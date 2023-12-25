package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	PostgresDSN      string `mapstructure:"POSTGRES_DSN"`
	ImagesFolderName string `mapstructure:"IMAGES_FOLDER_NAME"`
}

func MustParseConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		panic(err)
	}

	return &config
}
