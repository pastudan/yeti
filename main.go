package main

import (
	"fmt"
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type CoinbaseMessage struct {
	CommandType string `json:"type"`
}

type SubscribeMessage struct {
	CoinbaseMessage
	ProductID string `json:"product_id"`
}

type OrderMessage struct {
	CoinbaseMessage
	Price    string `json:"price"`
	Sequence int    `json:"sequence"`
}

type ReceivedOrderMessage struct {
	OrderMessage
}

type OpenOrderMessage struct {
	OrderMessage
}

type DoneOrderMessage struct {
	OrderMessage
}

type MatchOrderMessage struct {
	OrderMessage
}

type ChangeOrderMessage struct {
	OrderMessage
}

type ErrorMessage struct {
	CoinbaseMessage
	Message string `json:"message"`
}

func main() {
	log.Print("Connecting to the Coinbase Exchange real-time websocket feed...")

	statsdclient, err := statsd.NewClient("statsd.jacobgreenleaf.com:8125", "coinbase.")

	coinbase_url_raw := "wss://ws-feed.exchange.coinbase.com"
	coinbase_headers := http.Header{}
	coinbase_headers.Set("Origin", "http://www.jacobgreenleaf.com")

	ws_dialer := websocket.Dialer{}

	ws, _, err := ws_dialer.Dial(coinbase_url_raw, coinbase_headers)

	if err != nil {
		log.Fatalf("Error upgrading coinbase exchange feed connection to WebSocket: %s", err.Error())
	}

	log.Print("Subscribing to BTC-USD")

	subscribe_msg := &SubscribeMessage{CoinbaseMessage: CoinbaseMessage{CommandType: "subscribe"}, ProductID: "BTC-USD"}

	if err := ws.WriteJSON(subscribe_msg); err != nil {
		log.Fatalf("Error writing subscribe message: %s", err.Error())
	}

	log.Printf("Subscribed. Reading...")

	for {
		coinbase_msg := &CoinbaseMessage{}

		if err := ws.ReadJSON(coinbase_msg); err != nil {
			log.Printf("Error reading from websocket: %s", err.Error())
			break
		}

		fmt.Printf("Received: %s.\n", coinbase_msg.CommandType)
		statsdclient.Inc(coinbase_msg.CommandType, 1, 1)
	}

	log.Printf("Exiting...")
}
