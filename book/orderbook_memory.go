package book

import "fmt"
import "sort"
import "time"

type OrderHistory struct {
	Mutations     []OrderMutation
	FirstVersion  *StatefulOrder
	LatestVersion *StatefulOrder
}

func (h *OrderHistory) String() string {
	return fmt.Sprintf("<OrderHistory of %s>", h.FirstVersion.String())
}

type InMemoryOrderBook struct {
	Book map[OrderID]OrderHistory
}

func (m *InMemoryOrderBook) String() string {
	return fmt.Sprintf("<InMemoryOrderBook with %d orders>", len(m.Book))
}

func NewInMemoryOrderBook() (b *InMemoryOrderBook) {
	bk := make(map[OrderID]OrderHistory)
	return &InMemoryOrderBook{Book: bk}
}

func (book *InMemoryOrderBook) applyMutations(order StatefulOrder, muts []OrderMutation) *StatefulOrder {
	latest_time := muts[0].GetTime()
	// Re-apply all the mutations
	for _, mutation := range muts {
		porder, err := mutation.Apply(&order)

		if err != nil {
			continue
		}

		order = *porder

		if !latest_time.Before(mutation.GetTime()) {
			latest_time = mutation.GetTime()
		}
	}

	order.LastMutation = latest_time

	return &order
}

func (book *InMemoryOrderBook) GetOrder(id OrderID) (*StatefulOrder, error) {
	history, ok := book.Book[id]
	if !ok {
		return nil, errOrderDoesNotExist
	} else {
		return history.LatestVersion, nil
	}
}

// GetOrderVersion returns the order with all mutations less than or equal to t applied.
func (book *InMemoryOrderBook) GetOrderVersion(id OrderID, t time.Time) (*StatefulOrder, error) {
	history, ok := book.Book[id]

	if !ok {
		return nil, errOrderDoesNotExist
	}

	order := *history.FirstVersion
	muts := make([]OrderMutation, 0)
	for _, mut := range history.Mutations {
		if mut.GetTime().Before(t) || mut.GetTime().Equal(t) {
			muts = append(muts, mut)
		}
	}
	sort.Sort(OrderMutationByTime(muts))

	if len(muts) > 0 {
		order = *book.applyMutations(order, muts)
	}

	return &order, nil
}

func (book *InMemoryOrderBook) PlaceOrder(order Order, size int64, t time.Time) (err error) {
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
		Mutations:     []OrderMutation{&OrderStateChange{State: STATE_PENDING, Time: t}},
		FirstVersion:  sorder,
		LatestVersion: sorder,
	}

	book.Book[order.ID] = history

	return nil
}

func (book *InMemoryOrderBook) MutateOrder(id OrderID, muts []OrderMutation) error {
	history, ok := book.Book[id]

	if !ok {
		return errOrderDoesNotExist
	}

	if len(muts) == 0 {
		return nil
	}

	// Copy the first version of the order
	// To be precise, we might want to copy the Makers which is a []OrderID
	// but it probably doesn't matter much
	order := *history.FirstVersion

	history.Mutations = append(history.Mutations, muts...)
	sort.Sort(OrderMutationByTime(history.Mutations))

	order = *book.applyMutations(order, history.Mutations)
	history.LatestVersion = &order
	book.Book[id] = history

	return nil
}
