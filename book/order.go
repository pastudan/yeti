package book

import "time"

var (
	SIDE_BUY  = "buy"
	SIDE_SELL = "sell"
)

type Order struct {
	ID    string
	Price int64
	Size  int64
	Side  string
}

type MakeOrder struct {
	Order
	Time time.Time
}

type OpenOrder struct {
	OrderID string
	Time    time.Time
}

type VoidOrder struct {
	OrderID string
	Time    time.Time
}

type MutateOrderSize struct {
	OrderID string
	OldSize int64
	NewSize int64
	Time    time.Time
}

type MatchOrder struct {
	MakerOrderID string
	TakerOrderID string
	TradeID      int64
	Time         time.Time
}
