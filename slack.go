package main

import (
	"fmt"
	"github.com/golang/parkingbot/models"
	"github.com/nlopes/slack"
	"log"
	"regexp"
	"strings"
)

const (
	// action is used for slack attament action.
	actionSelect  = "select"
	actionParking = "parking"
	actionCancel  = "cancel"

	typeButton = "button"

	size = 5
)

type Slack struct {
	slackClient *slack.Client
	rtm         *slack.RTM
	db          *models.DB
	botID       string
	channelID   string
}

// ListenAndResponse listens slack events and response
// particular messages. It replies by slack message button.
func (s *Slack) ListenAndResponse() {

	// Start listening slack events
	go s.rtm.ManageConnection()

	// Handle slack events
	for msg := range s.rtm.IncomingEvents {
		fmt.Println("Event Received:", msg.Type)
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if err := s.handleMessageEvent(ev); err != nil {
				log.Printf("[ERROR] Failed to handle message: %s", err)
			}
		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			break

		default:
			// Take no action
		}
	}
}

// handleMesageEvent handles message events.
func (s *Slack) handleMessageEvent(ev *slack.MessageEvent) error {

	//// Only response in specific channel. Ignore else.
	//if ev.Channel != s.channelID {
	//	log.Printf("%s %s", ev.Channel, ev.Msg.Text)
	//	return nil
	//}

	//// Only response mention to bot. Ignore else.
	//if !strings.HasPrefix(ev.Msg.Text, fmt.Sprintf("<@%s> ", s.botID)) {
	//	return nil
	//}

	text := ev.Text
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	matchedWhere, _ := regexp.MatchString("where", text)
	matchedRegister, _ := regexp.MatchString("register", text)
	matchedGoOut, _ := regexp.MatchString("go out", text)

	if ev.User != s.rtm.GetInfo().User.ID && matchedWhere {

		spaces, err := s.getSpaces()
		if err != nil {
			return err
		}
		_, _, err = s.slackClient.PostMessage(ev.Channel, s.where(spaces))
		if err != nil {
			return err
		}
	} else if ev.User != s.rtm.GetInfo().User.ID && matchedRegister {

	} else if ev.User != s.rtm.GetInfo().User.ID && matchedGoOut {

	} else if ev.User != "" {
		s.rtm.SendMessage(s.rtm.NewOutgoingMessage("I'm so sorry, I don't understand what you say, I'm learning!! Thanks.", ev.Channel))
	}

	return nil
}

func (s *Slack) getSpaces() ([]slack.AttachmentAction, error) {
	spc, err := s.db.AllSpaces()
	if err != nil {
		return nil, err
	}

	spaces := make([]slack.AttachmentAction, 0)
	for _, sp := range spc {
		spaces = append(spaces, slack.AttachmentAction{
			Name:  actionSelect,
			Text:  sp.NumberSpace,
			Type:  typeButton,
			Value: sp.NumberSpace,
		})
	}
	return spaces, nil
}

func (s *Slack) where(spaces []slack.AttachmentAction) slack.MsgOption {
	// value is passed to message handler when request is approved.
	attachment := slack.Attachment{
		Text:       fmt.Sprintf("The available parking spaces are: %d", len(spaces)),
		CallbackID: "parkingSpaces",
		Actions:    spaces[:size],
	}

	attachments := make([]slack.Attachment, 0)
	attachments = append(attachments, attachment)
	for i := size; i < len(spaces); i += size {
		var attachmentActions []slack.AttachmentAction
		if len(spaces[i:]) >= size {
			attachmentActions = spaces[i : i+size]
		} else {
			attachmentActions = spaces[i:]
		}

		attachments = append(attachments, slack.Attachment{
			CallbackID: "parkingSpaces",
			Actions:    attachmentActions,
		})
	}

	return slack.MsgOptionAttachments(attachments...)
}
