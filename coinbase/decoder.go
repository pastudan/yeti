package coinbase

import "encoding/json"

var (
	MESSAGE_OPEN = "open"
	MESSAGE_RECEIVED = "received"
	MESSAGE_MATCH = "match"
	MESSAGE_CHANGE = "change"
	MESSAGE_DONE = "done"
	MESSAGE_ERROR = "error"
)

func Decode(json []byte) []*OrderBookCommand {
	data := json.Unmarshal(json, 
}
