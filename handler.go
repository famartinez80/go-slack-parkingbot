package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang/parkingbot/models"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/nlopes/slack"
)

// interactionHandler handles interactive message response.
type interactionHandler struct {
	slack             *Slack
	verificationToken string
}

func (h interactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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
		log.Printf("[ERROR] Failed to decode json message from slack: %s", jsonStr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Only accept message from slack with valid token
	if message.Token != h.verificationToken {
		log.Printf("[ERROR] Invalid token: %s", message.Token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	action := message.ActionCallback.AttachmentActions[0]
	switch action.Name {
	case actionSelect:

		space := &models.Spaces{}
		err := toStruct(action.Value, space)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		attachment := slack.Attachment{
			Text:       fmt.Sprintf("Can you park in parking lot number %s ?", strings.Title(space.NumberSpace)),
			CallbackID: "confirmParking",
			Actions: []slack.AttachmentAction{
				{
					Name:  actionParking,
					Text:  "Yes",
					Type:  typeButton,
					Value: action.Value,
				},
				{
					Name: actionCancel,
					Text: "No",
					Type: typeButton,
				},
			},
		}

		options := slack.MsgOptionAttachments(attachment)

		_, _, err = h.slack.slackClient.PostMessage(message.Channel.ID, options)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	case actionParking:

		space, err := h.slack.db.SpaceByUser(message.User.ID)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		if space != nil {
			text := fmt.Sprintf("You already have the parking lot %s assigned. :neutral_face:", space.NumberSpace)
			h.slack.responseMessage(text, message.Channel.ID)
		} else {
			space = &models.Spaces{}
			err = toStruct(action.Value, space)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
			}

			space.IdUser = &message.User.ID
			space.Available = 0
			err = h.slack.db.UpdateSpace(space)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				h.slack.responseMessage(":+1:", message.Channel.ID)
			}
		}

		return

	case actionConfirmGoOut:

		space := &models.Spaces{}
		err := toStruct(action.Value, space)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		space.IdUser = nil
		space.Available = 1
		err = h.slack.db.UpdateSpace(space)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			text := fmt.Sprintf("The parking lot %s has been released, Thank you!", space.NumberSpace)
			h.slack.responseMessage(text, message.Channel.ID)
		}

		return

	case actionCancel:

		text := fmt.Sprintf(":x: @%s canceled the request", message.User.Name)
		h.slack.responseMessage(text, message.Channel.ID)
		return
	default:

		text := "I'm so sorry, I don't understand what you say, I'm learning!! Thanks."
		h.slack.responseMessage(text, message.Channel.ID)
		return
	}
}

func toStruct(s string, i interface{}) error {
	err := json.Unmarshal([]byte(s), i)
	if err != nil {
		return err
	}
	return nil
}
