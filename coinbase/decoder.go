package coinbase

import "encoding/json"
import "github.com/jacobgreenleaf/yeti/book"
import "strconv"
import "log"
import "time"
import "errors"
import "bytes"
import "fmt"

const (
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

var (
	errStaleCommand = errors.New("Order sequence is older than the book sequence.")
)

type CoinbaseOrderBookCommandBatch struct {
	Commands []book.OrderBookCommand
	Sequence int64
}

func (b *CoinbaseOrderBookCommandBatch) Apply(book book.OrderBook) error {
	for _, cmd := range b.Commands {
		if err := cmd.Apply(book); err != nil {
			return err
		}
	}

	return nil
}

func DecodeRealtimeEvent(rawMsg []byte) *CoinbaseOrderBookCommandBatch {
	var coinbaseEvent map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(rawMsg))
	decoder.UseNumber()
	err := decoder.Decode(&coinbaseEvent)

	if err != nil {
		log.Fatalf("Error decoding coinbase JSON: %s", err.Error())
		return nil
	}

	cmds := make([]book.OrderBookCommand, 0)

	coinbaseType, _ := coinbaseEvent["type"].(string)

	if "error" == coinbaseType {
		log.Printf("Received coinbase error: %s", coinbaseEvent["message"].(string))
		return nil
	}

	coinbaseTime, err := time.Parse(time.RFC3339Nano, coinbaseEvent["time"].(string))
	if err != nil {
		log.Fatalf("Failed to parse timestamp %s", coinbaseEvent["time"].(string))
	}

	coinbaseSide := coinbaseEvent["side"].(string)
	coinbasePrice, err := strconv.ParseFloat(coinbaseEvent["price"].(string), 64)

	if err != nil {
		log.Fatalf("Failed to parse float price %s: %s", coinbaseEvent["price"].(string), err.Error())
		return nil
	}

	coinbasePriceCents := int64(coinbasePrice * 100)
	coinbaseSequenceNumber, err := coinbaseEvent["sequence"].(json.Number).Int64()
	if err != nil {
		log.Fatalf("Failed to parse sequence number %s: %s", coinbaseEvent["sequence"].(json.Number), err.Error())
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
			log.Fatalf("Failed to parse float size: %s", coinbaseEvent["remaining_size"].(string))
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
		makerId := book.OrderID(coinbaseEvent["maker_order_id"].(string))
		takerId := book.OrderID(coinbaseEvent["taker_order_id"].(string))

		coinbaseSize, err := strconv.ParseFloat(coinbaseEvent["size"].(string), 64)
		if err != nil {
			log.Fatalf("Failed to parse float size: %s", coinbaseEvent["size"].(string))
			return nil
		}
		coinbaseSizeSatoshi := int64(coinbaseSize * float64(SATOSHI))
		tradeId, err := coinbaseEvent["trade_id"].(json.Number).Int64()
		if err != nil {
			log.Fatalf("Failed to parse trade id %s: %s", coinbaseEvent["trade_id"].(json.Number), err.Error())
			return nil
		}

		takerMuts := []book.OrderMutation{&book.OrderMatchMutation{
			TradeID:  tradeId,
			Size:     coinbaseSizeSatoshi,
			WasMaker: false,
			MakerID:  makerId,
			Time:     coinbaseTime,
		}}

		cmdTaker := &book.OrderBookMutationCommand{
			ID:        takerId,
			Mutations: takerMuts,
		}

		cmds = append(cmds, cmdTaker)

		makerMuts := []book.OrderMutation{&book.OrderMatchMutation{
			TradeID:  tradeId,
			Size:     coinbaseSizeSatoshi,
			WasMaker: true,
			Time:     coinbaseTime,
		}}

		cmdMaker := &book.OrderBookMutationCommand{
			ID:        makerId,
			Mutations: makerMuts,
		}

		cmds = append(cmds, cmdMaker)

		break
	case MESSAGE_CHANGE:
		coinbaseSize, err := strconv.ParseFloat(coinbaseEvent["new_size"].(string), 64)
		if err != nil {
			log.Fatalf("Failed to parse float size: %s", coinbaseEvent["new_size"].(string))
			return nil
		}
		coinbaseSizeSatoshi := int64(coinbaseSize * float64(SATOSHI))

		muts := []book.OrderMutation{&book.OrderSizeMutation{
			NewSize: coinbaseSizeSatoshi,
			Time:    coinbaseTime,
		}}

		cmds = append(cmds, &book.OrderBookMutationCommand{
			ID:        book.OrderID(coinbaseEvent["order_id"].(string)),
			Mutations: muts,
		})
		break
	case MESSAGE_ERROR:
		break
	}

	return &CoinbaseOrderBookCommandBatch{
		Commands: cmds,
		Sequence: coinbaseSequenceNumber,
	}
}

func DecodeRESTOrderBook(rawMsg []byte) (coinbaseSequenceNumber int64, batch *CoinbaseOrderBookCommandBatch, err error) {
	var msg map[string]interface{}

	decoder := json.NewDecoder(bytes.NewReader(rawMsg))
	decoder.UseNumber()

	err = decoder.Decode(&msg)

	coinbaseSequenceNumber, err = msg["sequence"].(json.Number).Int64()
	if err != nil {
		return 0, nil, fmt.Errorf("Error parsing sequence number: %s", err.Error())
	}

	bids := msg["bids"].([]interface{})
	asks := msg["asks"].([]interface{})

	decodeCoinbaseOrder := func(side string, coinbaseOrder []interface{}) (*book.OrderBookPlacementCommand, error) {
		priceDollars, err := strconv.ParseFloat(coinbaseOrder[0].(string), 64)
		if err != nil {
			return nil, fmt.Errorf("Error parsing order price %s: %s", coinbaseOrder[0].(string), err.Error())
		}
		priceCents := int64(priceDollars * 100)
		sizeBitcoins, err := strconv.ParseFloat(coinbaseOrder[1].(string), 64)
		if err != nil {
			return nil, fmt.Errorf("Error parsing order size %s: %s", coinbaseOrder[1].(string), err.Error())
		}
		sizeSatoshi := int64(sizeBitcoins * SATOSHI)
		orderId := coinbaseOrder[2].(string)
		order := book.Order{
			ID:    book.OrderID(orderId),
			Side:  side,
			Price: priceCents,
		}
		cmd := &book.OrderBookPlacementCommand{
			Size:  sizeSatoshi,
			Order: order,
			// Time intentionally left zeroed since Coinbase doesn't tell us when it was placed
		}
		return cmd, nil
	}

	cmds := make([]book.OrderBookCommand, 0, len(bids)+len(asks))

	for _, bid := range bids {
		orderCmd, err := decodeCoinbaseOrder(book.SIDE_BUY, bid.([]interface{}))
		if err == nil {
			cmds = append(cmds, orderCmd)
		} else {
			return 0, nil, err
		}
	}
	for _, ask := range asks {
		orderCmd, err := decodeCoinbaseOrder(book.SIDE_SELL, ask.([]interface{}))
		if err == nil {
			cmds = append(cmds, orderCmd)
		} else {
			return 0, nil, err
		}
	}

	batch = &CoinbaseOrderBookCommandBatch{
		Commands: cmds,
		Sequence: coinbaseSequenceNumber,
	}

	return coinbaseSequenceNumber, batch, nil
}
