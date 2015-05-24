package book

import "time"
import "fmt"

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

func (o *Order) String() string {
	return fmt.Sprintf("<%s order at price %d; id=%s>", o.Side, o.Price, o.ID)
}

type StatefulOrder struct {
	Order
	Size         int64
	State        string
	LastMutation time.Time
	Makers       []OrderID
}

func (o *StatefulOrder) String() string {
	return fmt.Sprintf("<%s StatefulOrder at price %d for %d units; last changed at %s; id=%s>", o.Side, o.Price, o.Size, o.LastMutation, o.ID)
}
