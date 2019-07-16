package parkingbot

import (
	"encoding/json"
	"fmt"
	"github.com/go-slack-parkingbot/model"
	"github.com/nlopes/slack"
	"log"
	"regexp"
	"strings"
)

var cache = make(map[string]*model.User, 0)

const (
	// action is used for Slack attachment action.
	ActionSelect   = "select"
	ActionParking  = "parking"
	ActionCancel   = "cancel"
	ActionGoOut    = "goOut"
	ActionRegistry = "registry"

	CallBackRegistry = "registry"

	TypeButton = "button"

	Size = 5
)

type Slack struct {
	SlackClient *slack.Client
	RTM         *slack.RTM
	DB          *model.DB
	BotID       string
	ChannelID   string
}

// ListenAndResponse listens Slack events and response
// particular messages. It replies by Slack message button.
func (s *Slack) ListenAndResponse() {

	// Start listening Slack events
	go s.RTM.ManageConnection()

	// Handle Slack events
	for msg := range s.RTM.IncomingEvents {
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

// handleMessageEvent handles message events.
func (s *Slack) handleMessageEvent(ev *slack.MessageEvent) error {

	//// Only response in specific channel. Ignore else.
	//if ev.Channel != s.ChannelID {
	//	log.Printf("%s %s", ev.Channel, ev.Msg.Text)
	//	return nil
	//}

	//// Only response mention to bot. Ignore else.
	//if !strings.HasPrefix(ev.Msg.Text, fmt.Sprintf("<@%s> ", s.BotID)) {
	//	return nil
	//}

	user, err := s.getUser(ev.User)
	if err != nil {
		return err
	}

	text := ev.Text
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	matchedWhere, _ := regexp.MatchString("where", text)
	matchedGoOut, _ := regexp.MatchString("go out", text)

	if ev.User != "" && user == nil {

		_, _, err = s.SlackClient.PostMessage(ev.Channel, s.registry())
		if err != nil {
			return err
		}

	} else if ev.User != s.RTM.GetInfo().User.ID && matchedWhere {

		spaces, err := s.getSpaces()
		if err != nil {
			return err
		}
		_, _, err = s.SlackClient.PostMessage(ev.Channel, s.where(spaces))
		if err != nil {
			return err
		}

	} else if ev.User != s.RTM.GetInfo().User.ID && matchedGoOut {

		space, err := s.DB.SpaceByUser(ev.User)
		if err != nil {
			return err
		}

		if space != nil {
			_, _, err = s.SlackClient.PostMessage(ev.Channel, s.goOut(*space))
			if err != nil {
				return err
			}
		} else {
			s.ResponseMessage("You don't have a parking lot assigned.", ev.Channel)
		}

	} else if ev.User != "" {
		s.ResponseMessage("I'm so sorry, I don't understand what you say, I'm learning!! Thanks.", ev.Channel)
	}

	return nil
}

func (s *Slack) getSpaces() ([]slack.AttachmentAction, error) {
	spc, err := s.DB.AllSpaces()
	if err != nil {
		return nil, err
	}

	spaces := make([]slack.AttachmentAction, 0)
	for _, sp := range spc {
		spaces = append(spaces, slack.AttachmentAction{
			Name:  ActionSelect,
			Text:  sp.NumberSpace,
			Type:  TypeButton,
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
		Actions:    spaces[:Size],
	}

	attachments := make([]slack.Attachment, 0)
	attachments = append(attachments, attachment)
	for i := Size; i < len(spaces); i += Size {
		var attachmentActions []slack.AttachmentAction
		if len(spaces[i:]) >= Size {
			attachmentActions = spaces[i : i+Size]
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

func (s *Slack) goOut(space model.Space) slack.MsgOption {
	attachment := slack.Attachment{
		Text:       fmt.Sprintf("Do you want to release parking lot number %s ?", strings.Title(space.NumberSpace)),
		CallbackID: "goOutParking",
		Actions: []slack.AttachmentAction{
			{
				Name:  ActionGoOut,
				Text:  "Yes",
				Type:  TypeButton,
				Value: toJson(space),
			},
			{
				Name: ActionCancel,
				Text: "No",
				Type: TypeButton,
			},
		},
	}

	return slack.MsgOptionAttachments(attachment)
}

func (s *Slack) parking(space model.Space, value string) slack.MsgOption {
	attachment := slack.Attachment{
		Text:       fmt.Sprintf("Can you park in parking lot number %s ?", strings.Title(space.NumberSpace)),
		CallbackID: "confirmParking",
		Actions: []slack.AttachmentAction{
			{
				Name:  ActionParking,
				Text:  "Yes",
				Type:  TypeButton,
				Value: value,

			},
			{
				Name: ActionCancel,
				Text: "No",
				Type: TypeButton,
			},
		},
	}

	return slack.MsgOptionAttachments(attachment)
}

func (s *Slack) registry() slack.MsgOption {

	attachment := slack.Attachment{
		Text:       "You have not registered yet, do you want to register now?",
		CallbackID: CallBackRegistry,
		Actions: []slack.AttachmentAction{
			{
				Name: ActionRegistry,
				Text: "Yes",
				Type: TypeButton,
			},
			{
				Name: ActionCancel,
				Text: "No",
				Type: TypeButton,
			},
		},
	}

	return slack.MsgOptionAttachments(attachment)
}

func toJson(i interface{}) string {
	b, _ := json.Marshal(i)
	return string(b)
}

func toStruct(s string, i interface{}) error {
	err := json.Unmarshal([]byte(s), i)
	if err != nil {
		return err
	}
	return nil
}

func (s *Slack) ResponseMessage(text, channel string) {
	s.RTM.SendMessage(s.RTM.NewOutgoingMessage(text, channel))
}

func (s *Slack) getUser(id string) (*model.User, error) {
	var usr *model.User
	var err error

	if usr = cache[id]; usr == nil {
		usr, err = s.DB.FindUser(id)
		cache[id] = usr
	}

	return usr, err
}