package main

import (
	"bitbucket.org/jacobgreenleaf/yeti/book"
	"bitbucket.org/jacobgreenleaf/yeti/coinbase"
	//"container/list"
	"encoding/json"
	"time"
	//"github.com/cactus/go-statsd-client/statsd"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
)

func main() {
	var err error

	log.Printf("Connecting to Coinbase Exchange real-time API...")

	coinbase_url_raw := "wss://ws-feed.exchange.coinbase.com"
	coinbase_headers := http.Header{}
	coinbase_headers.Set("Origin", "http://www.jacobgreenleaf.com")
	coinbase_headers.Set("User-Agent", "Yeti <jacob@jacobgreenleaf.com>")
	ws_dialer := websocket.Dialer{}
	ws, _, err := ws_dialer.Dial(coinbase_url_raw, coinbase_headers)

	if err != nil {
		log.Fatalf("Error upgrading coinbase exchange feed connection to WebSocket: %s", err.Error())
	}

	log.Printf("Connected. Subscribing to BTC-USD...")

	ws.WriteMessage(websocket.TextMessage, []byte(`
		{
			"type": "subscribe",
			"product_id": "BTC-USD"
		}
	`))

	log.Printf("Synchronizing order book...")

	orderBook := book.NewInMemoryOrderBook()

	for {
		var reader io.Reader
		_, reader, err := ws.NextReader()

		if err != nil {
			log.Printf("Error getting next reader from websocket: %s", err.Error())
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

			cmds := coinbase.Decode(rawBytes)

			if cmds == nil || len(cmds) == 0 {
				break
			}

			for _, cmd := range cmds {
				err = cmd.Apply(orderBook)
				if err != nil {
					log.Printf("Failed to apply order book command: %s", err.Error())
				}
			}

			priceLevels := orderBook.GetPriceLevels()
			centsInPlay := book.CalculateTotalCentsInPlayInMemory(orderBook, time.Now())
			openOrders := book.CalculateNumberOfOpenOrdersInMemory(orderBook, time.Now())

			log.Printf("There are %d open orders at %d price levels and %d dollars in play", openOrders, len(priceLevels), centsInPlay/int64(100*coinbase.SATOSHI))
		}
	}

	log.Printf("Exiting...")
}
