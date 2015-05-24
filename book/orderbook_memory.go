package book

//import "fmt"
//import "time"

type OrderHistory struct {
	Mutations     []OrderMutation
	FirstVersion  *StatefulOrder
	LatestVersion *StatefulOrder
}

type InMemoryOrderBook struct {
	Book map[OrderID]OrderHistory
}

func NewInMemoryOrderBook() (b *InMemoryOrderBook) {
	bk := make(map[OrderID]OrderHistory)
	return &InMemoryOrderBook{Book: bk}
}

func (book *InMemoryOrderBook) GetOrder(id OrderID) (sorder *StatefulOrder, err error) {
	history, ok := book.Book[id]
	if !ok {
		return nil, errOrderDoesNotExist
	} else {
		return history.LatestVersion, nil
	}
}

func (book *InMemoryOrderBook) PlaceOrder(order Order, size int64) (err error) {
	_, ok := book.Book[order.ID]

	if ok {
		return errOrderAlreadyExists
	}

	sorder := &StatefulOrder{
		Order:  order,
		State:  STATE_PENDING,
		Size:   size,
		Makers: nil,
	}

	history := OrderHistory{
		Mutations:     make([]OrderMutation, 0),
		FirstVersion:  sorder,
		LatestVersion: sorder,
	}

	book.Book[order.ID] = history

	return nil
}

func (book *InMemoryOrderBook) MutateOrder(id OrderID, muts []OrderMutation) (err error) {
	history, ok := book.Book[id]

	if !ok {
		return errOrderDoesNotExist
	}

	if len(muts) == 0 {
		return nil
	}

	history.Mutations = append(history.Mutations, muts...)

	order := *history.FirstVersion
	// To be precise, we might want to copy the Makers which is a []OrderID
	// but it probably doesn't matter much

	latest_time := muts[0].GetTime()
	// Re-apply all the mutations
	for _, mutation := range history.Mutations {
		order = *mutation.Apply(&order)
		if latest_time.Before(mutation.GetTime()) {
			latest_time = mutation.GetTime()
		}
	}

	order.LastMutation = latest_time
	history.LatestVersion = &order
	book.Book[id] = history

	return nil
}

/*
type OrderBook interface {
	PlaceOrder(Order, size int64) error
	MutateOrder(OrderID, []OrderMutation, time.Time) error
	VoidOrder(OrderID) error

	GetOrder(OrderID) (StatefulOrder, error)
	GetOrderVersion(OrderID, time.Time) (StatefulOrder, error)
}
*/
