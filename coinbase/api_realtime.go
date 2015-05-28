package coinbase

import "github.com/gorilla/websocket"
import "net/http"
import "io"
import "encoding/json"

const (
	COINBASE_WEBSOCKET_URL = "wss://ws-feed.exchange.coinbase.com"
)

type OrderBookCommandFeed struct {
	Feed   chan *CoinbaseOrderBookCommandBatch
	socket *websocket.Conn
}

func ConnectRealtimeFeed(bufLen int) (*OrderBookCommandFeed, error) {
	headers := http.Header{}
	headers.Set("Origin", "http://www.jacobgreenleaf.com")
	headers.Set("User-Agent", "Yeti <jacob@jacobgreenleaf.com>")
	dialer := websocket.Dialer{}
	socket, _, err := dialer.Dial(COINBASE_WEBSOCKET_URL, headers)

	if err != nil {
		return nil, err
	} else {
		feed := make(chan *CoinbaseOrderBookCommandBatch, bufLen)
		cmdFeed := &OrderBookCommandFeed{
			Feed:   feed,
			socket: socket,
		}
		return cmdFeed, nil
	}
}

func (feed *OrderBookCommandFeed) ReadForever() {
	for {
		var reader io.Reader
		_, reader, err := feed.socket.NextReader()

		if err != nil {
			// log.Printf("Error getting next reader from websocket: %s", err.Error())
			continue
		}

		decoder := json.NewDecoder(reader)

		for {
			var rawOrder interface{}
			decoder.Decode(&rawOrder)

			if rawOrder == nil {
				break
			}

			rawBytes, _ := json.Marshal(rawOrder)

			batch := DecodeRealtimeEvent(rawBytes)

			if batch != nil && len(batch.Commands) > 0 {
				feed.Feed <- batch
			}
		}
	}
}

func (feed *OrderBookCommandFeed) Subscribe(product string) {
	type msg struct {
		Type      string `json:"type"`
		ProductID string `json:"product_id"`
	}

	subscribeMsg := msg{Type: "subscribe", ProductID: product}

	subscribeMsgBytes, _ := json.Marshal(subscribeMsg)

	feed.socket.WriteMessage(websocket.TextMessage, subscribeMsgBytes)
}
