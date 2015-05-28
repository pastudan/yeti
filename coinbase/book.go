package coinbase

import "github.com/jacobgreenleaf/yeti/book"

// The Available mutex represents the code's knowledge of whether the order book is stale.
//
// When out of order events come through the web socket, the update routine
// will grab a write lock, preventing anyone after who cares from reading a known stale version. It
// will then buffer for a fixed amount of events before either panicking or finally
// getting a complete ordered set of events, which it will then write and release the write
// lock. People who try to grab read locks during that time will be blocked because they
// don't want to read a known stale version.
type CoinbaseOrderBook struct {
	Book      book.OrderBook
	Available *sync.RWMutex
	feed      coinbase.OrderBookCommandFeed
}

func Bootstrap() (*CoinbaseOrderBook, error) {

}

// It is recomended to spawn this in a goroutine.
func (b *CoinbaseOrderBook) MaintainForever() {

}
