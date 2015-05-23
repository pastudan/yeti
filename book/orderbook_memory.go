package book

import "container/list"
import "errors"
import "time"

type OrderHistory struct {
	Mutations     *list.List
	FirstVersion  *StatefulOrder
	LatestVersion *StatefulOrder
}

type InMemoryOrderBook struct {
	Book      map[string]*OrderHistory
	LastTrade *Order // Maker side
}

func NewInMemoryOrderBook() (b *InMemoryOrderBook) {
	bk := make(map[string]*OrderHistory)
	return &InMemoryOrderBook{Book: bk, LastTrade: nil}
}

/*
type OrderBook interface {
	PlaceOrder(Order, size int64) error
	MutateOrders([]OrderMutation) error
	VoidOrder(OrderID) error

	GetLastTrade() (taker Order, maker []Order, err error)
	GetOrder(OrderID) (StatefulOrder, error)
	GetOrderVersion(OrderID, time.Time) (StatefulOrder, error)
}
*/
