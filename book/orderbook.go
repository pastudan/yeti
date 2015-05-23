package book

import "container/list"
import "errors"
import "time"

var (
	errOrderAlreadyExists = errors.New("Order already exists.")
	errOrderDoesNotExist  = errors.New("Order does not exist.")
	errOrderIsAlreadyOpen = errors.New("Order is already open.")
)

type OrderBook interface {
	PlaceOrder(Order, size int64) error
	MutateOrders([]OrderMutation) error
	VoidOrder(OrderID) error

	GetLastTrade() (taker Order, maker []Order, err error)
	GetOrder(OrderID) (StatefulOrder, error)
	GetOrderVersion(OrderID, time.Time) (StatefulOrder, error)
}

type OrderMutation struct {
	ID   OrderID
	Time time.Time
}

type OrderStateChange struct {
	OrderMutation
	State string
}

type OrderSizeChange struct {
	OrderMutation
	OldSize int64
	NewSize int64
}

type OrderMatch struct {
	OrderMutation
	TradeID string
	OtherID OrderID
}
