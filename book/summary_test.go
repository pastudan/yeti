package book

import "testing"
import "time"

func TestCalculateCentsInPlayInMemory(t *testing.T) {
	book := NewInMemoryOrderBook()
	orderOne := Order{ID: "foobar", Price: 100, Side: SIDE_BUY}
	orderTwo := Order{ID: "bazbar", Price: 100, Side: SIDE_SELL}
	book.PlaceOrder(orderOne, 10, time.Unix(0, 0))
	book.PlaceOrder(orderTwo, 10, time.Unix(0, 0))

	centsInPlayAtTimeZero := CalculateTotalCentsInPlayInMemory(book, time.Unix(0, 0))
	if centsInPlayAtTimeZero != 0 {
		t.Fatalf("Expected number of cents in play at t=0 to be zero, instead %d", centsInPlayAtTimeZero)
	}

	book.MutateOrder("foobar", []OrderMutation{&OrderStateMutation{
		State: STATE_OPEN,
		Time:  time.Unix(1, 0),
	}})

	centsInPlayAtTimeZero = CalculateTotalCentsInPlayInMemory(book, time.Unix(0, 0))
	if centsInPlayAtTimeZero != 0 {
		t.Fatalf("Expected number of cents in play at t=0, after mutation at t=1, to be zero, instead %d", centsInPlayAtTimeZero)
	}

	centsInPlayAtTimeOne := CalculateTotalCentsInPlayInMemory(book, time.Unix(1, 0))
	if centsInPlayAtTimeOne != 1000 {
		t.Fatalf("Expected number of cents in play at t=1 to be 1000, instead %d", centsInPlayAtTimeOne)
	}

	book.MutateOrder("bazbar", []OrderMutation{&OrderStateMutation{
		State: STATE_OPEN,
		Time:  time.Unix(1, 0),
	}})

	centsInPlayAtTimeOne = CalculateTotalCentsInPlayInMemory(book, time.Unix(1, 0))
	if centsInPlayAtTimeOne != 2000 {
		t.Fatalf("Expected number of cents in play at t=1 after mutation to be 2000, instead %d", centsInPlayAtTimeOne)
	}
	centsInPlayAtTimeZero = CalculateTotalCentsInPlayInMemory(book, time.Unix(0, 0))
	if centsInPlayAtTimeZero != 0 {
		t.Fatalf("Expected number of cents in play at t=0 after mutation to be 0, instead %d", centsInPlayAtTimeZero)
	}

	book.MutateOrder("bazbar", []OrderMutation{&OrderSizeMutation{
		NewSize: 5,
		Time:    time.Unix(2, 0),
	}})

	centsInPlayAtTimeTwo := CalculateTotalCentsInPlayInMemory(book, time.Unix(2, 0))
	if centsInPlayAtTimeTwo != 1500 {
		t.Fatalf("Expected number of cents in play at t=2 to be 1500, instead %d", centsInPlayAtTimeTwo)
	}
}

func TestCalculateNumberOfOpenOrders(t *testing.T) {
	book := NewInMemoryOrderBook()
	orderOne := Order{ID: "foobar", Price: 100, Side: SIDE_BUY}
	orderTwo := Order{ID: "bazbar", Price: 100, Side: SIDE_SELL}
	book.PlaceOrder(orderOne, 10, time.Unix(0, 0))
	book.PlaceOrder(orderTwo, 10, time.Unix(0, 0))

	openOrders := CalculateNumberOfOpenOrdersInMemory(book, time.Unix(0, 0))
	if openOrders != 0 {
		t.Fatalf("Expected number of open orders at t=0 to be 0, instead %d", openOrders)
	}

	book.MutateOrder("foobar", []OrderMutation{&OrderStateMutation{
		State: STATE_OPEN,
		Time:  time.Unix(1, 0),
	}})

	openOrders = CalculateNumberOfOpenOrdersInMemory(book, time.Unix(0, 0))
	if openOrders != 0 {
		t.Fatalf("Expected number of open orders at t=0 after mutation to be 0, instead %d", openOrders)
	}
	openOrders = CalculateNumberOfOpenOrdersInMemory(book, time.Unix(1, 0))
	if openOrders != 1 {
		t.Fatalf("Expected number of open orders at t=1 to be 1, instead %d", openOrders)
	}

	book.MutateOrder("bazbar", []OrderMutation{&OrderStateMutation{
		State: STATE_OPEN,
		Time:  time.Unix(2, 0),
	}})

	openOrders = CalculateNumberOfOpenOrdersInMemory(book, time.Unix(1, 0))
	if openOrders != 1 {
		t.Fatalf("Expected number of open orders at t=1 after mutation to be 1, instead %d", openOrders)
	}
	openOrders = CalculateNumberOfOpenOrdersInMemory(book, time.Unix(2, 0))
	if openOrders != 2 {
		t.Fatalf("Expected number of open orders at t=2 after mutation to be 2, instead %d", openOrders)
	}
}

func TestCalculateBidMedianAsk(t *testing.T) {
	var bid, median, ask, spread int64

	book := NewInMemoryOrderBook()
	orderOne := Order{ID: "foobar", Price: 100, Side: SIDE_BUY}
	orderTwo := Order{ID: "bazbar", Price: 110, Side: SIDE_SELL}
	book.PlaceOrder(orderOne, 10, time.Unix(0, 0))
	book.PlaceOrder(orderTwo, 10, time.Unix(0, 0))

	bid, median, ask, spread = CalculateBidMedianAskSpreadInMemory(book, time.Unix(0, 0))
	if bid != -1 {
		t.Fatalf("Expected bid at t=0 to be -1, instead %d", bid)
	}
	if median != -1 {
		t.Fatalf("Expected median at t=0 to be -1, instead %d", median)
	}
	if ask != -1 {
		t.Fatalf("Expected ask at t=0 to be -1, instead %d", ask)
	}
	if spread != 0 {
		t.Fatalf("Expected spread at t=0 to be -1, instead %d", spread)
	}

	book.MutateOrder("foobar", []OrderMutation{&OrderStateMutation{
		State: STATE_OPEN,
		Time:  time.Unix(1, 0),
	}})

	book.MutateOrder("bazbar", []OrderMutation{&OrderStateMutation{
		State: STATE_OPEN,
		Time:  time.Unix(1, 0),
	}})

	bid, median, ask, spread = CalculateBidMedianAskSpreadInMemory(book, time.Unix(1, 0))
	if bid != 100 {
		t.Fatalf("Expected bid at t=1 to be 100, instead %d", bid)
	}
	if ask != 110 {
		t.Fatalf("Expected ask at t=1 to be 110, instead %d", ask)
	}
	if median != 105 {
		t.Fatalf("Expected median at t=1 to be 105, instead %d", median)
	}
	if spread != 10 {
		t.Fatalf("Expected spread at t=1 to be 10, instead %d", spread)
	}

}
