package main

import (
	"fmt"
	"github.com/spf13/viper"
)

type config struct {
	Port              string `mapstructure:"port"`
	BotToken          string `mapstructure:"bot_token"`
	BotID             string `mapstructure:"bot_id"`
	VerificationToken string `mapstructure:"verification_token"`
	ChannelID         string `mapstructure:"channel_id"`
	DataSource        string `mapstructure:"data_source"`
}

func initConfig() *config {

	viper.AutomaticEnv()
	viper.SetEnvPrefix("parking")
	// Port is server port to be listened.
	viper.SetDefault("PORT", "")
	// BotToken is bot user token to access to slack API.
	viper.SetDefault("BOT_TOKEN", "")
	// BotID is bot user ID.
	viper.SetDefault("BOT_ID", "")
	// VerificationToken is used to validate interactive messages from slack.
	viper.SetDefault("VERIFICATION_TOKEN", "")
	// ChannelID is slack channel ID where bot is working.
	viper.SetDefault("CHANNEL_ID", "")
	// DataSource database mysql
	viper.SetDefault("DATA_SOURCE", "")

	conf := &config{}
	err := viper.Unmarshal(conf)
	if err != nil {
		fmt.Printf("Unable to decode into config struct, %v", err)
	}

	return conf
}
