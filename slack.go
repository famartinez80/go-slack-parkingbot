package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang/parkingbot/models"
	"github.com/nlopes/slack"
	"log"
	"regexp"
	"strings"
)

const (
	// action is used for slack attament action.
	actionSelect       = "select"
	actionParking      = "parking"
	actionCancel       = "cancel"
	actionGoOut        = "goOut"
	actionConfirmGoOut = "confirmGoOut"

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

		space, err := s.db.SpaceByUser(ev.User)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			s.responseMessage("It's been an error, I'm sorry", ev.Channel)
		}
		if space != nil {

			attachment := slack.Attachment{
				Text:       fmt.Sprintf("Do you want to release parking lot number %s ?", strings.Title(space.NumberSpace)),
				CallbackID: "confirmParking",
				Actions: []slack.AttachmentAction{
					{
						Name:  actionConfirmGoOut,
						Text:  "Yes",
						Type:  typeButton,
						Value: toJson(space),
					},
					{
						Name: actionCancel,
						Text: "No",
						Type: typeButton,
					},
				},
			}

			options := slack.MsgOptionAttachments(attachment)
			_, _, err = s.slackClient.PostMessage(ev.Channel, options)
			if err != nil {
				return err
			}

		} else {
			s.responseMessage("You don't have a parking lot assigned. :sad_face:", ev.Channel)
		}

	} else if ev.User != "" {
		s.responseMessage("I'm so sorry, I don't understand what you say, I'm learning!! Thanks.", ev.Channel)
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
			Value: toJson(sp),
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

func toJson(i interface{}) string {
	b, _ := json.Marshal(i)
	return string(b)
}
func (s *Slack) responseMessage(text, channel string) {
	s.rtm.SendMessage(s.rtm.NewOutgoingMessage(text, channel))
}
