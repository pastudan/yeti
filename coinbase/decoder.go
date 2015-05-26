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
	REASON_FILLED    = "filled"
	REASON_CANCELLED = "cancelled"
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

		if err != nil {
			log.Fatalf("Failed to parse float size: %s", coinbaseEvent["size"].(string))
			return nil
		}

		coinbaseSizeSatoshi := int64(coinbaseSize * float64(SATOSHI))

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

		coinbaseSize, err := strconv.ParseFloat(coinbaseEvent["remaining_size"].(string), 64)

		if err != nil {
			log.Fatalf("Failed to parse float size: %s", coinbaseEvent["size"].(string))
			return nil
		}

		coinbaseSizeSatoshi := int64(coinbaseSize * float64(SATOSHI))

		muts := make([]book.OrderMutation, 0, 2)
		muts = append(muts, &book.OrderSizeMutation{
			NewSize: coinbaseSizeSatoshi,
			Time:    coinbaseTime,
		})
		muts = append(muts, &book.OrderStateMutation{
			State: book.STATE_OPEN,
			Time:  coinbaseTime,
		})

		cmds = append(cmds, &book.OrderBookMutationCommand{
			ID:        book.OrderID(coinbaseEvent["order_id"].(string)),
			Mutations: muts,
		})
		break
	case MESSAGE_DONE:
		reason := coinbaseEvent["reason"].(string)

		coinbaseSize, err := strconv.ParseFloat(coinbaseEvent["remaining_size"].(string), 64)
		if err != nil {
			log.Fatalf("Failed to parse float size: %s", coinbaseEvent["size"].(string))
			return nil
		}
		coinbaseSizeSatoshi := int64(coinbaseSize * float64(SATOSHI))

		var coinbaseState string

		if REASON_FILLED == reason {
			coinbaseState = book.STATE_FILLED
		} else if REASON_CANCELLED == reason {
			coinbaseState = book.STATE_VOID
		}

		muts := make([]book.OrderMutation, 0, 2)
		muts = append(muts, &book.OrderSizeMutation{
			NewSize: coinbaseSizeSatoshi,
			Time:    coinbaseTime,
		})
		muts = append(muts, &book.OrderStateMutation{
			State: coinbaseState,
			Time:  coinbaseTime,
		})

		cmds = append(cmds, &book.OrderBookMutationCommand{
			ID:        book.OrderID(coinbaseEvent["order_id"].(string)),
			Mutations: muts,
		})
		break
	case MESSAGE_MATCH:
		maker_id := coinbaseEvent["maker_order_id"].(string)
		taker_id := coinbaseEvent["taker_order_id"].(string)

		taker_muts := []book.OrderMutation{&book.OrderMatchMutation{
			TradeID  string
			Size     int64
			WasMaker bool
			MakerID  OrderID
			Time     time.Time
		}}

		cmd_taker := &bok.OrderBookMutationCommand{
			ID:        book.OrderID(taker_id),
			Mutations: taker_muts,
		}

		break
	case MESSAGE_CHANGE:
		break
	case MESSAGE_ERROR:
		break
	}

	return cmds
}
