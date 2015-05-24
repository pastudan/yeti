package book

import "errors"
import "time"

var (
	errOrderAlreadyExists = errors.New("Order already exists.")
	errOrderDoesNotExist  = errors.New("Order does not exist.")
	errOrderIsAlreadyOpen = errors.New("Order is already open.")
)

type OrderBook interface {
	PlaceOrder(Order, size int64) error
	MutateOrder(OrderID, []OrderMutation, time.Time) error
	VoidOrder(OrderID) error

	GetOrder(OrderID) (StatefulOrder, error)
	GetOrderVersion(OrderID, time.Time) (StatefulOrder, error)
}

type OrderMutation interface {
	Apply(*StatefulOrder) *StatefulOrder
	GetTime() time.Time
}

type OrderStateChange struct {
	State string
	Time  time.Time
}

func (m *OrderStateChange) GetTime() time.Time {
	return m.Time
}

func (m *OrderStateChange) Apply(s *StatefulOrder) *StatefulOrder {
	new_order := *s // copy
	new_order.State = m.State
	return &new_order
}

type OrderSizeChange struct {
	NewSize int64
	Time    time.Time
}

func (m *OrderSizeChange) Apply(s *StatefulOrder) *StatefulOrder {
	new_order := *s // copy
	new_order.Size = m.NewSize
	return &new_order
}

func (m *OrderSizeChange) GetTime() time.Time {
	return m.Time
}

type OrderMatch struct {
	TradeID  string
	Size     int64
	WasMaker bool
	MakerID  OrderID
	Time     time.Time
}

func (m *OrderMatch) Apply(s *StatefulOrder) *StatefulOrder {
	new_order := *s // copy
	new_order.Size -= m.Size

	if !m.WasMaker {
		new_order.Makers = append(new_order.Makers, m.MakerID)
	}

	return &new_order
}

func (m *OrderMatch) GetTime() time.Time {
	return m.Time
}
