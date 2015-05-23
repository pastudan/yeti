package book

import "container/list"

type OrderBook struct {
	book      map[string]*list.List
	LastTrade *Order
}

type OrderBookAction interface {
}
