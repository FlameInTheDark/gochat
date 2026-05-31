package botgateway

import "encoding/json"

type DeliveryMessage struct {
	SessionIDs []string        `json:"session_ids"`
	Event      json.RawMessage `json:"event"`
}
