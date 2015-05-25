package coinbase

import "encoding/json"
import "bitbucket.org/jacobgreenleaf/yeti/book"
import "strconv"
import "log"
import "time"

var (
	MESSAGE_OPEN     = "open"
	MESSAGE_RECEIVED = "received"
	MESSAGE_MATCH    = "match"
	MESSAGE_CHANGE   = "change"
	MESSAGE_DONE     = "done"
	MESSAGE_ERROR    = "error"
	SATOSHI          = 100000000
)

func Decode(rawMsg []byte) []book.OrderBookCommand {
	var coinbaseEvent map[string]interface{}

	err := json.Unmarshal(rawMsg, &coinbaseEvent)

	if err != nil {
		log.Fatalf("Error decoding coinbase JSON: %s", err.Error())
		return nil
	}

	cmds := make([]book.OrderBookCommand, 0)

	coinbaseType, _ := coinbaseEvent["type"].(string)

	if "error" == coinbaseType {
		log.Fatalf("Received coinbase error: %s", coinbaseEvent["message"].(string))
		return nil
	}

	coinbaseTime, err := time.Parse(time.RFC3339Nano, coinbaseEvent["time"].(string))
	if err != nil {
		log.Fatalf("Failed to parse timestamp %s", coinbaseEvent["time"].(string))
	}

	coinbaseSide := coinbaseEvent["side"].(string)
	coinbasePrice, err := strconv.ParseFloat(coinbaseEvent["price"].(string), 64)
	coinbasePriceCents := int64(coinbasePrice * 100)

	if err != nil {
		log.Fatalf("Failed to parse float price %s", coinbaseEvent["price"].(string))
		return nil
	}

	switch coinbaseType {
	case MESSAGE_RECEIVED:

		coinbaseSize, err := strconv.ParseFloat(coinbaseEvent["size"].(string), 64)
		coinbaseSizeSatoshi := int64(coinbaseSize * float64(SATOSHI))

		if err != nil {
			log.Fatalf("Failed to parse float size: %s", coinbaseEvent["size"].(string))
			return nil
		}

		cmds = append(cmds, &book.OrderBookPlacementCommand{
			Order: book.Order{
				ID:    book.OrderID(coinbaseEvent["order_id"].(string)),
				Price: coinbasePriceCents,
				Side:  coinbaseSide,
			},
			Size: coinbaseSizeSatoshi,
			Time: coinbaseTime,
		})
		break
	case MESSAGE_OPEN:
		break
	case MESSAGE_MATCH:
		break
	case MESSAGE_CHANGE:
		break
	case MESSAGE_DONE:
		break
	case MESSAGE_ERROR:
		break
	}

	return cmds
}
