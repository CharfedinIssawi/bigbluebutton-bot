package bot

import (
	api "api"
	"errors"
	"net/http"
	"time"

	ddp "ddp"
)

type StatusType string

const (
	DISCONNECTING StatusType = "disconnecting"
	DISCONNECTED  StatusType = "disconnected"
	CONNECTING    StatusType = "connecting"
	CONNECTED     StatusType = "connected"
	RECONNECTING  StatusType = "reconnecting"
)

// This is for all events that are in "event_....go" files
type eventDDPHandler struct {
	client *Client
}

// Will be emited by ddpClient
func (e *eventDDPHandler) CollectionUpdate(collection string, operation string, id string, doc ddp.Update) {
	// "redirect" to the event handler
	switch collection {
	case "group-chat-msg":
		e.client.updateGroupChatMsg(collection, operation, id, doc)
	default:
		// do nothing
		return
	}
}

// Client represents a BigBlueButton client connection. The BigBlueButton client establish a BigBlueButton
// session and acts as a message pump for other tools.
type Client struct {
	// Status is the current connection status of the client
	Status StatusType

	// BBB-urls the client is connected to
	ClientURL   string
	ClientWSURL string
	PadURL   	string
	PadWSURL 	string
	ApiURL      string
	apiSecret   string
	// to make api requests to the BBB-server
	API *api.ApiRequest

	ddpClient *ddp.Client

	// events will store all the functions executed on certain events. (events["OnStatus"][]func(StatusType))
	events          map[string][]interface{}
	eventDDPHandler *eventDDPHandler

	// after validateAuthToken there are the following informations
	// ConnectionID string `json:"connectionId"`
	// MeetingID    string `json:"meetingId"` // internal meetingID
	// UserID       string `json:"userId"`

	// after join there are the following informations
	JoinURL           string
	SessionCookie     []*http.Cookie
	InternalUserID    string
	AuthToken         string
	SessionToken      string
	InternalMeetingID string
}

func NewClient(clientURL string, clientWSURL string, padURL string, padWSURL string, apiURL string, apiSecret string) (*Client, error) {
	api, err := api.NewRequest(apiURL, apiSecret, api.SHA256)
	if err != nil {
		return nil, err
	}

	ddpClient := ddp.NewClient(clientWSURL, clientURL)

	c := &Client{
		Status: DISCONNECTED,

		ClientURL:   clientURL,
		ClientWSURL: clientWSURL,
		PadURL:   	 padURL,
		PadWSURL: 	 padWSURL,
		ApiURL:      apiURL,
		apiSecret:   apiSecret,

		ddpClient: ddpClient,

		API: api,

		events:          make(map[string][]interface{}),
		eventDDPHandler: nil,
	}

	c.eventDDPHandler = &eventDDPHandler{
		client: c,
	}

	return c, nil
}

// Join a meeting
func (c *Client) Join(meetingID string, userName string, moderator bool) error {
	joinURL, coockie, internalUserID, authToken, sessionToken, internalMeetingID, err := c.API.Join(meetingID, userName, moderator)
	c.JoinURL = joinURL
	c.SessionCookie = coockie
	c.InternalUserID = internalUserID
	c.AuthToken = authToken
	c.SessionToken = sessionToken
	c.InternalMeetingID = internalMeetingID

	if err != nil {
		return err
	}

	err = c.ddpClient.Connect()
	if err != nil {
		return err
	}

	err = c.ddpClient.Sub("current-user")
	if err != nil {
		return errors.New("could sub current-user")
	}

	// Call the validateAuthToken method with the userID, authToken, and userName
	_, err = c.ddpClient.Call("validateAuthToken", internalMeetingID, internalUserID, authToken, internalUserID)
	if err != nil {
		return errors.New("could not validateAuthToken")
	}

	return nil
}

// Leave the joined meeting
func (c *Client) Leave() error {
	// If not connected, return an error
	if c.Status != CONNECTED {
		// If is connecting retry 5 times
		if c.Status == CONNECTING {
			i := 0
			for i < 5 {
				if c.Status == CONNECTED {
					c.Leave()
				}
				time.Sleep(time.Second * 1)
				i += 1
			}
		}
		return errors.New("Client is in no meeting. First Join a meeting with: client.Join(meetingID string, userName string, moderator bool)")
	}

	c.ddpClient.Call("userLeftMeeting")
	c.ddpClient.Call("setExitReason", "logout")
	// c.ddpClient.UnSubscribe("from all subs")

	c.ddpClient.Close()

	c.ddpClient = nil

	return nil
}
