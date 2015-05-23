package book

import "time"

var (
	SIDE_BUY  = "buy"
	SIDE_SELL = "sell"
)

const (
	STATE_PENDING = "pending"
	STATE_OPEN    = "open"
	STATE_FILLED  = "filled"
	STATE_VOID    = "void"
)

const (
	MATCH_TYPE_TAKER = "taker"
	MATCH_TYPE_MAKER = "maker"
)

type OrderID string

type Order struct {
	ID    OrderID
	Price int64
	Side  string
}

type StatefulOrder struct {
	Order
	State  string
	Makers []OrderID
}
