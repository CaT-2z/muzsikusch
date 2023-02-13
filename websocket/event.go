package websocket

import (
	"encoding/json"
	"fmt"
)

// Event is the Messages sent over the websocket
// Used to differ between different actions
type Event struct {
	// Type is the message type sent
	Type string `json:"Type"`
	// Payload is the data Based on the Type
	Payload json.RawMessage `json:"Payload"`
}

type EventManager struct {
	events []EventHandler
}

type EventHandler struct {
	name      string
	eventFunc EventFunc
}

func NewEventManager() *EventManager {
	return &EventManager{
		events: []EventHandler{
			{
				name:      "pause",
				eventFunc: PauseEventHandler,
			},
			{
				name:      "seek",
				eventFunc: SeekEventHandler,
			},
			{
				name:      "volume",
				eventFunc: VolumeEventHandler,
			},
		},
	}
}

// EventHandler is a function signature that is used to affect messages on the socket and triggered
// depending on the type
type EventFunc func(event Event, c *Client) error

type PauseEventPayload struct {
	TimeStamp float32 `json:"TimeStamp"`
}

func PauseEventHandler(event Event, c *Client) error {
	return nil
}

func SeekEventHandler(event Event, c *Client) error {
	return nil
}

func VolumeEventHandler(event Event, c *Client) error {
	return nil
}

func (e *EventManager) HandleEvent(event Event, c *Client) error {
	for _, ev := range e.events {
		if ev.name == event.Type {
			return ev.eventFunc(event, c)
		}
	}
	fmt.Println("No event type recognised: ", event.Type)
	return nil
}
