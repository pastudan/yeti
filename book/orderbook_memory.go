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
	Book               map[OrderID]*OrderHistory
	PriceLevels        map[int64][]*OrderHistory
	History            []*OrderHistory
	LatestMutationTime time.Time
}

func (m *InMemoryOrderBook) String() string {
	return fmt.Sprintf("<InMemoryOrderBook with %d orders>", len(m.Book))
}

func NewInMemoryOrderBook() (b *InMemoryOrderBook) {
	bk := make(map[OrderID]*OrderHistory)
	prices := make(map[int64][]*OrderHistory)
	history := make([]*OrderHistory, 0)
	return &InMemoryOrderBook{bk, prices, history, *new(time.Time)}
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

	order.LatestMutationTime = latest_time

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

	history := &OrderHistory{
		Mutations:     []OrderMutation{&OrderStateMutation{State: STATE_PENDING, Time: t}},
		FirstVersion:  sorder,
		LatestVersion: sorder,
	}

	book.Book[order.ID] = history

	_, ok = book.PriceLevels[order.Price]
	if !ok {
		book.PriceLevels[order.Price] = make([]*OrderHistory, 0)
	}

	book.PriceLevels[order.Price] = append(book.PriceLevels[order.Price], history)
	book.History = append(book.History, history)

	if book.LatestMutationTime.Before(t) {
		book.LatestMutationTime = t
	}

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

	if book.LatestMutationTime.Before(order.LatestMutationTime) {
		book.LatestMutationTime = order.LatestMutationTime
	}

	return nil
}

func (book *InMemoryOrderBook) GetPriceLevel(level int64) []*StatefulOrder {
	return book.GetPriceLevelVersion(level, book.LatestMutationTime)
}

// Retrieve all of the orders at a given price level
func (book *InMemoryOrderBook) GetPriceLevelVersion(level int64, t time.Time) []*StatefulOrder {
	histories, ok := book.PriceLevels[level]

	if !ok {
		return nil
	}

	prices := make([]*StatefulOrder, 0, len(histories))

	// Filter only open orders
	for _, history := range histories {
		var order *StatefulOrder = nil

		// We might be able to just use the latest version and avoid (possibly) expensively
		// replaying the mutations if all the updates happened before t (latest)
		if history.LatestVersion.LatestMutationTime.After(t) {
			// Drats, we have to apply only the mutations that occurred before or at t
			order, _ = book.GetOrderVersion(history.LatestVersion.ID, t)
		} else {
			order = history.LatestVersion
		}

		if order.State == STATE_OPEN || order.State == STATE_PENDING {
			prices = append(prices, order)
		}
	}

	return prices
}

func (book *InMemoryOrderBook) GetPriceLevels() []int64 {
	keys := make([]int64, 0, len(book.PriceLevels))

	for price_level := range book.PriceLevels {
		keys = append(keys, price_level)
	}

	return keys
}

// Vacuuming removes all of the voided or filled orders
func (book *InMemoryOrderBook) Vacuum() {
	for order_id, history := range book.Book {
		if history.LatestVersion.State == STATE_FILLED || history.LatestVersion.State == STATE_VOID {
			delete(book.Book, order_id)
		}
	}
}
