package book

import "errors"
import "time"
import "fmt"

var (
	errOrderAlreadyExists        = errors.New("Order already exists.")
	errOrderDoesNotExist         = errors.New("Order does not exist.")
	errOrderIsAlreadyOpen        = errors.New("Order is already open.")
	errOrderSizeMutationTooLarge = errors.New("Order size is too large; quantity would go below zero")
)

type OrderBook interface {
	PlaceOrder(Order, size int64, t time.Time) error
	MutateOrder(OrderID, []OrderMutation) error

	GetOrder(OrderID) (StatefulOrder, error)
	GetOrderVersion(OrderID, time.Time) (StatefulOrder, error)
	GetPriceLevel(int64) []*StatefulOrder
}

type OrderMutation interface {
	Apply(*StatefulOrder) (*StatefulOrder, error)
	GetTime() time.Time
}

type OrderMutationByTime []OrderMutation

func (a OrderMutationByTime) Len() int           { return len(a) }
func (a OrderMutationByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a OrderMutationByTime) Less(i, j int) bool { return a[i].GetTime().Before(a[j].GetTime()) }

type OrderStateChange struct {
	State string
	Time  time.Time
}

func (m *OrderStateChange) String() string {
	return fmt.Sprintf("<OrderStateChange to '%s' at %s>", m.State, m.Time.String())
}

func (m *OrderStateChange) Apply(s *StatefulOrder) (*StatefulOrder, error) {
	new_order := *s // copy
	new_order.State = m.State
	return &new_order, nil
}

func (m *OrderStateChange) GetTime() time.Time {
	return m.Time
}

type OrderSizeChange struct {
	NewSize int64
	Time    time.Time
}

func (m *OrderSizeChange) String() string {
	return fmt.Sprintf("<OrderSizeChange to %d units at %s>", m.NewSize, m.Time.String())
}

func (m *OrderSizeChange) Apply(s *StatefulOrder) (*StatefulOrder, error) {
	new_order := *s // copy
	new_order.Size = m.NewSize
	return &new_order, nil
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

func (m *OrderMatch) String() string {
	return fmt.Sprintf("<OrderMatch of %d units at %s; trade id=%s>", m.Size, m.Time.String(), m.TradeID)
}

func (m *OrderMatch) Apply(s *StatefulOrder) (*StatefulOrder, error) {
	new_order := *s // copy

	if s.State != STATE_OPEN {
		return &new_order, nil
	}

	if s.Size-m.Size < 0 {
		return nil, errOrderSizeMutationTooLarge
	}

	new_order.Size -= m.Size

	if !m.WasMaker {
		new_order.Makers = append(new_order.Makers, m.MakerID)
	}

	if new_order.Size == 0 {
		new_order.State = STATE_FILLED
	}

	return &new_order, nil
}

func (m *OrderMatch) GetTime() time.Time {
	return m.Time
}
