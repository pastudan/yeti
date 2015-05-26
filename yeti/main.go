package main

import (
	"github.com/jacobgreenleaf/yeti/book"
	"github.com/jacobgreenleaf/yeti/coinbase"
	//"container/list"
	"time"
	//"github.com/cactus/go-statsd-client/statsd"
	"log"
)

func main() {
	var err error

	log.Printf("Connecting to Coinbase Exchange real-time API...")

	feed, err := coinbase.ConnectRealtimeFeed(1000)

	if err != nil {
		log.Fatalf("Error upgrading coinbase exchange feed connection to WebSocket: %s", err.Error())
	}

	log.Printf("Connected. Subscribing to BTC-USD...")

	feed.Subscribe("BTC-USD")

	log.Printf("Synchronizing order book...")

	orderBook := book.NewInMemoryOrderBook()

	var cmds []coinbase.CoinbaseOrderBookCommand = nil

	go feed.ReadForever()

	for {
		select {
		case cmds = <-feed.Feed:
			for _, cmd := range cmds {
				err = cmd.Command.Apply(orderBook)
				if err != nil {
					log.Printf("Failed to apply order book command: %s", err.Error())
				}
			}

			openOrders := book.CalculateNumberOfOpenOrdersInMemory(orderBook, time.Now())
			bid, median, ask, spread := book.CalculateBidMedianAskSpreadInMemory(orderBook, time.Now())

			log.Printf("There are %d open orders. Bid: %d\tMed: %d\tAsk: %d\tSpread: %d", openOrders, bid, median, ask, spread)

			orderBook.Vacuum()
		}
	}

	log.Printf("Exiting...")
}
