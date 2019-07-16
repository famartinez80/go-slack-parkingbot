package parkingbot

import (
	"encoding/json"
	"fmt"
	"github.com/go-slack-parkingbot/model"
	"github.com/nlopes/slack"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

// InteractionHandler handles interactive message response.
type InteractionHandler struct {
	Slack             *Slack
	VerificationToken string
}

func (h InteractionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		log.Printf("[ERROR] Invalid method: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonStr, err := url.QueryUnescape(string(buf)[8:])
	if err != nil {
		log.Printf("[ERROR] Failed to unespace request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var message slack.InteractionCallback
	if err := json.Unmarshal([]byte(jsonStr), &message); err != nil {
		log.Printf("[ERROR] Failed to decode json message from Slack: %s", jsonStr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Only accept message from Slack with valid token
	if message.Token != h.VerificationToken {
		log.Printf("[ERROR] Invalid token: %s", message.Token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if len(message.ActionCallback.AttachmentActions) > 0 {
		action := message.ActionCallback.AttachmentActions[0]
		switch action.Name {
		case ActionSelect:

			space := &model.Space{}
			err := toStruct(action.Value, space)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
			}

			_, _, err = h.Slack.SlackClient.PostMessage(message.Channel.ID, h.Slack.parking(*space, action.Value))
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
			}

			return
		case ActionParking:

			space, err := h.Slack.DB.SpaceByUser(message.User.ID)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
			}

			if space != nil {
				text := fmt.Sprintf("You already have the parking lot %s assigned. :neutral_face:", space.NumberSpace)
				h.Slack.ResponseMessage(text, message.Channel.ID)
			} else {
				space = &model.Space{}
				err = toStruct(action.Value, space)
				if err != nil {
					fmt.Printf("Error: %s\n", err)
					w.WriteHeader(http.StatusInternalServerError)
				}

				space.IdUser = &message.User.ID
				space.Available = 0
				err = h.Slack.DB.UpdateSpace(space)
				if err != nil {
					fmt.Printf("Error: %s\n", err)
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					h.Slack.ResponseMessage(":+1:", message.Channel.ID)
				}
			}

			return

		case ActionGoOut:

			space := &model.Space{}
			err := toStruct(action.Value, space)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
			}

			space.IdUser = nil
			space.Available = 1
			err = h.Slack.DB.UpdateSpace(space)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				text := fmt.Sprintf("The parking lot %s has been released, Thank you!", space.NumberSpace)
				h.Slack.ResponseMessage(text, message.Channel.ID)
			}

			return

		case ActionRegistry:

			dialog := slack.Dialog{}
			dialog.TriggerID = message.TriggerID
			dialog.CallbackID = message.CallbackID
			dialog.Title = "New user?"
			dialog.Elements = []slack.DialogElement{
				slack.TextInputElement{
					DialogInput: slack.DialogInput{
						Label: "Phone number",
						Name:  "number",
						Type:  slack.InputTypeText,
					},
					MaxLength: 10,
					MinLength: 7,
					Subtype:   slack.InputSubtypeTel,
				},
			}

			err := h.Slack.RTM.OpenDialog(message.TriggerID, dialog)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
			}

			return

		case ActionCancel:

			text := fmt.Sprintf(":x: @%s canceled the request", message.User.Name)
			h.Slack.ResponseMessage(text, message.Channel.ID)
			return
		default:

			text := "I'm so sorry, I don't understand what you say, I'm learning!! Thanks."
			h.Slack.ResponseMessage(text, message.Channel.ID)
			return
		}
	} else {
		switch message.CallbackID {
		case CallBackRegistry:

			usr := model.User{
				ID:        message.User.ID,
				ChannelID: message.Channel.ID,
				Mobile:    message.DialogSubmissionCallback.Submission["number"],
			}

			err = h.Slack.DB.CreateUser(usr)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				h.Slack.ResponseMessage("User has been created.", message.Channel.ID)
			}

			return

		default:
		}
	}
}
