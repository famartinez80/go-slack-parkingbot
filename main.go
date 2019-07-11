package main

import (
	"github.com/golang/parkingbot/models"
	"github.com/nlopes/slack"
	"log"
	"net/http"
	"os"
)

// https://api.slack.com/slack-apps
// https://api.slack.com/internal-integrations
func main() {
	os.Exit(_main(os.Args[1:]))
}

func _main(args []string) int {
	// Listening slack event and response
	log.Printf("[INFO] Start slack event listening")

	config, err := initConfig()
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return 1
	}
	client := slack.New(config.BotToken)
	rtm := client.NewRTM()

	db, err := models.InitDB(config.DataSource)
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return 1
	}

	slackListener := &Slack{
		slackClient: client,
		rtm:         rtm,
		db:          db,
		botID:       config.BotID,
		channelID:   config.ChannelID,
	}

	go slackListener.ListenAndResponse()

	// Register handler to receive interactive message
	// responses from slack (kicked by user action)
	http.Handle("/interaction", interactionHandler{
		slack:             slackListener,
		verificationToken: config.VerificationToken,
	})

	log.Printf("[INFO] Server listening on :%s", config.Port)
	if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
		log.Printf("[ERROR] %s", err)
		return 1
	}

	return 0

}
