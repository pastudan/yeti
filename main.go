package main

import (
	//	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/gorilla/websocket"
	//"github.com/sdming/goh"
	"io"
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
	Time     string `json:"time"`
	OrderID  string `json:"order_id"`
	Side     string `json:"side"`
}

type ReceivedOrderMessage struct {
	OrderMessage
	Size string `json:"size"`
}

type OpenOrderMessage struct {
	OrderMessage
	RemainingSize string `json:"remaining_size"`
}

type DoneOrderMessage struct {
	OrderMessage
	Reason        string `json:"reason"`
	RemainingSize string `json:"remaining_size"`
}

type ChangeOrderMessage struct {
	OrderMessage
	NewSize string `json:"new_size"`
	OldSize string `json:"old_size"`
}

type MatchOrderMessage struct {
	CoinbaseMessage
	Price        string `json:"price"`
	Sequence     int    `json:"sequence"`
	Side         string `json:"side"`
	MakerOrderID string `json:"maker_order_id"`
	TakerOrderID string `json:"taker_order_id"`
	TradeID      int    `json:"trade_id"`
}

type ErrorMessage struct {
	CoinbaseMessage
	Message string `json:"message"`
}

func main() {
	log.Printf("Connecting to HBase")

	/*
		hbaseclient, err := goh.NewTcpClient("127.0.0.1:9090", goh.TBinaryProtocol, false)
		if err != nil {
			log.Fatalf("Error connecting to HBase: %s", err.Error())
		}

		if err = hbaseclient.Open(); err != nil {
			log.Fatalf("Error opening client to HBase: %s", err.Error())
		}

		defer hbaseclient.Close()
	*/

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
		var reader io.Reader

		_, reader, err := ws.NextReader()

		if err != nil {
			log.Printf("Error getting reader from websocket: %s", err.Error())
			break
		}

		decoder := json.NewDecoder(reader)
		for {
			var raw_order map[string]interface{}
			decoder.Decode(&raw_order)

			fmt.Printf("Received %s\n", raw_order)
			/*
				coinbase_msg := &CoinbaseMessage{}

				var order interface{}

				if "open" == coinbase_msg.CommandType {
					order = &OpenOrderMessage{}
				} else if "received" == coinbase_msg.CommandType {
					order = &ReceivedOrderMessage{}
				} else if "match" == coinbase_msg.CommandType {
					order = &MatchOrderMessage{}
				} else if "change" == coinbase_msg.CommandType {
					order = &ChangeOrderMessage{}
				} else if "done" == coinbase_msg.CommandType {
					order = &DoneOrderMessage{}
				} else if "error" == coinbase_msg.CommandType {
					order = &ErrorMessage{}
				} else {
					fmt.Printf("Unknown order type %s", coinbase_msg.CommandType)
				}

				fmt.Printf("Received: %s.\n", order)
			*/
			statsdclient.Inc("", 1, 1)
		}
	}

	log.Printf("Exiting...")
}
