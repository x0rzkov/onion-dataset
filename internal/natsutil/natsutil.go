package natsutil

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

// PublishJSON publish given message serialized in json with given subject
func PublishJSON(nc *nats.Conn, subject string, msg interface{}) error {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error while encoding message: %s", err)
	}

	return nc.Publish(subject, msgBytes)
}

// ReadJSON read given encoded json message and deserialize into into given structure
func ReadJSON(msg *nats.Msg, body interface{}) error {
	if err := json.Unmarshal(msg.Data, body); err != nil {
		return fmt.Errorf("error while decoding message: %s", err)
	}

	return nil
}
