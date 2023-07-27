package websocket

import (
	"encoding/json"
	"fmt"
	"muzsikusch/src/queue/entry"

	"github.com/jonhoo/go-events"
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
	name string
}

func NewEventManager() *EventManager {
	return &EventManager{
		events: []EventHandler{
			{
				name: "pause",
			},
			{
				name: "seek",
			},
			{
				name: "volume",
			},
		},
	}
}

// EventHandler is a function signature that is used to affect messages on the socket and triggered
// depending on the type
// type EventFunc func(event Event, c *Client) error

type PauseEventPayload struct {
	TimeStamp float32 `json:"TimeStamp"`
}

func CreatePauseEvent(timestamp float32) Event {
	return CreateEvent("pause", PauseEventPayload{TimeStamp: timestamp})
}

// func PauseEventHandler(event Event, c *Client) error {
// 	return nil
// }

type UnpauseEventPayload = PauseEventPayload

func CreateUnpauseEvent(timestamp float32) Event {
	return CreateEvent("unpause", UnpauseEventPayload{TimeStamp: timestamp})
}

type AppendEventPayload struct {
	Entry entry.Entry `json:"Entry"`
}

func CreateAppendEvent(entry entry.Entry) Event {
	return CreateEvent("append", AppendEventPayload{Entry: entry})
}

type PushEventPayload = AppendEventPayload

func CreatePushEvent(entry entry.Entry) Event {
	return CreateEvent("push", PushEventPayload{Entry: entry})
}

type TrackStartEventPayload = AppendEventPayload

func CreateTrackStartEvent(entry entry.Entry) Event {
	return CreateEvent("start", PushEventPayload{Entry: entry})
}

type RemoveEventPayload struct {
	UID string `json:"UID"`
}

func CreateRemoveEvent(UID string) Event {
	return CreateEvent("remove", RemoveEventPayload{UID: UID})
}

// func SeekEventHandler(event Event, c *Client) error {
// 	return nil
// }

// func VolumeEventHandler(event Event, c *Client) error {
// 	return nil
// }

func (e *EventManager) HandleEvent(event Event, c *Client) error {
	events.Announce(events.Event{Tag: event.Type, Data: event.Payload})
	return nil
}

func CreateEvent(Type string, v any) Event {
	js, err := json.Marshal(v)
	if err != nil {
		fmt.Println("Couldnt marshal object ", err)
		return Event{}
	}

	return Event{
		Type:    Type,
		Payload: js,
	}
}

// func (e *EventManager)SetEventHandler(string event, EventHandler func) bool {
// 	for _, ev := range e.events {
// 		if ev.name == event {
// 			ev.eventFunc = func;
// 			return true
// 		}
// 	}
// 	return false
// }
