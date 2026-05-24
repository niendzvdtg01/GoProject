package consumer

import (
	"encoding/json"

	"backend/package/event"
)

func decodeEvent(body []byte) (event.Event, error) {
	var evt event.Event
	if err := json.Unmarshal(body, &evt); err != nil {
		return event.Event{}, err
	}
	return evt, nil
}
