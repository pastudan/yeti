package book

import "testing"
import "time"

func TestPlacingOrders(t *testing.T) {
	book := NewInMemoryOrderBook()

	sorder, err := book.GetOrder("foobar")
	if err == nil {
		t.Fatal("Expected getting an non-existent order to return an error")
	}

	order := Order{ID: "foobar", Price: 100, Side: SIDE_BUY}
	book.PlaceOrder(order, 10, time.Unix(0, 0))

	sorder, err = book.GetOrder("foobar")
	if err != nil {
		t.Fatalf("Expected getting an order after placing it to return an order, instead got %s", err.Error())
	}
	if sorder.Order != order {
		t.Fatalf("Expected placed order %s to equal retrieved order %s", sorder.Order, order)
	}
	if sorder.State != STATE_PENDING {
		t.Fatalf("Expected just placed order %s to have pending state", sorder)
	}
	if sorder.Size != 10 {
		t.Fatalf("Expected order size %d to be 10", sorder.Size)
	}
}

func TestMutatingSingleOrder(t *testing.T) {
	book := NewInMemoryOrderBook()
	order := Order{ID: "foobar", Price: 100, Side: SIDE_BUY}
	book.PlaceOrder(order, 10, time.Unix(0, 0))

	mut := &OrderStateChange{
		State: STATE_OPEN,
	}
	errs := book.MutateOrder("foobar", []OrderMutation{mut})
	if errs != nil {
		t.Fatalf("Unexpected error mutating order book: %s", errs)
	}
	sorder, err := book.GetOrder("foobar")
	if sorder.State != STATE_OPEN {
		t.Fatalf("Mutation failed to apply. Expected state %s to be %s", sorder.State, STATE_OPEN)
	}

	mut = &OrderStateChange{
		State: STATE_OPEN,
	}
	errs = book.MutateOrder("bazbar", []OrderMutation{mut})
	if errs == nil {
		t.Fatal("Expected state mutation on non-existent order to be invalid")
	}

	sizemut := &OrderSizeChange{
		NewSize: 11,
	}
	errs = book.MutateOrder("foobar", []OrderMutation{sizemut})
	if errs != nil {
		t.Fatalf("Unexpected errors mutating order: %s", errs)
	}
	sorder, err = book.GetOrder("foobar")
	if err != nil {
		t.Fatalf("Failed to get mutated order: %s", err.Error())
	}
	if sorder.Size != 11 {
		t.Fatalf("Expected mutated order size %d to be 11", sorder.Size)
	}

	sizemut_new := &OrderSizeChange{
		NewSize: 20,
		Time:    time.Unix(1, 0),
	}
	sizemut_old := &OrderSizeChange{
		NewSize: 15,
		Time:    time.Unix(0, 0),
	}
	errs = book.MutateOrder("foobar", []OrderMutation{sizemut_new, sizemut_old})
	if errs != nil {
		t.Fatalf("Unexpected errors mutating order: %s", errs)
	}
	sorder, err = book.GetOrder("foobar")
	if err != nil {
		t.Fatalf("Failed to get mutated order: %s", err.Error())
	}
	if sorder.Size != 20 {
		t.Fatalf("Mutations failed to respect time ordering. Expected order size %d to be 20", sorder.Size)
	}

	match := &OrderMatch{
		TradeID:  "a",
		Size:     15,
		WasMaker: true,
		Time:     time.Unix(2, 0),
	}
	errs = book.MutateOrder("foobar", []OrderMutation{match})
	if errs != nil {
		t.Fatalf("Unexpected error mutating order: %s", errs)
	}
	sorder, err = book.GetOrder("foobar")
	if err != nil {
		t.Fatalf("Failed to get mutated order: %s", err.Error())
	}
	if sorder.Size != 5 {
		t.Fatalf("Expected a match of 15 units on a 20 unit order to result in 5 units, instead %d units remain", sorder.Size)
	}
	if sorder.State != STATE_OPEN {
		t.Fatalf("Expected partially filled order to still be open, instead %s", sorder.State)
	}

	match = &OrderMatch{
		TradeID:  "b",
		Size:     5,
		WasMaker: true,
		Time:     time.Unix(3, 0),
	}
	errs = book.MutateOrder("foobar", []OrderMutation{match})
	if errs != nil {
		t.Fatalf("Unexpected error mutating order: %s", errs)
	}
	sorder, err = book.GetOrder("foobar")
	if err != nil {
		t.Fatalf("Failed to get mutated order: %s", err.Error())
	}
	if sorder.Size != 0 {
		t.Fatalf("Expected a match of 5 units on a 5 unit order to result in 0 units, instead %d units remain", sorder.Size)
	}
	if sorder.State != STATE_FILLED {
		t.Fatalf("Expected fully filled order to be state filled, instead %s", sorder.State)
	}

	match = &OrderMatch{
		TradeID:  "c",
		Size:     1,
		WasMaker: true,
		Time:     time.Unix(4, 0),
	}
	errs = book.MutateOrder("foobar", []OrderMutation{match})
	sorder, err = book.GetOrder("foobar")
	if err != nil {
		t.Fatalf("Failed to get mutated order: %s", err.Error())
	}
	if sorder.Size != 0 {
		t.Fatalf("Expected an invalid match change on filled order to have size 0; instead size %d", sorder.Size)
	}
}

func TestVoidingOrder(t *testing.T) {
	book := NewInMemoryOrderBook()
	order := Order{ID: "foobar", Price: 100, Side: SIDE_SELL}
	book.PlaceOrder(order, 10, time.Unix(0, 0))

	mut := &OrderStateChange{
		State: STATE_OPEN,
	}
	errs := book.MutateOrder("foobar", []OrderMutation{mut})
	if errs != nil {
		t.Fatalf("Unexpected error mutating order book: %s", errs)
	}
	sorder, _ := book.GetOrder("foobar")
	if sorder.State != STATE_OPEN {
		t.Fatalf("Mutation failed to apply. Expected state %s to be %s", sorder.State, STATE_OPEN)
	}

	mut = &OrderStateChange{
		State: STATE_OPEN,
	}
	errs = book.MutateOrder("bazbar", []OrderMutation{mut})
	if errs == nil {
		t.Fatal("Expected state mutation on non-existent order to be invalid")
	}
}

func TestOrderVersions(t *testing.T) {
	book := NewInMemoryOrderBook()
	order := Order{ID: "foobar", Price: 100, Side: SIDE_BUY}
	book.PlaceOrder(order, 10, time.Unix(0, 0))

	mut := &OrderSizeChange{
		NewSize: 9,
		Time:    time.Unix(1, 0),
	}
	err := book.MutateOrder("foobar", []OrderMutation{mut})

	mut = &OrderSizeChange{
		NewSize: 5,
		Time:    time.Unix(2, 0),
	}
	err = book.MutateOrder("foobar", []OrderMutation{mut})

	sorderAtZero, err := book.GetOrderVersion("foobar", time.Unix(0, 0))
	if err != nil {
		t.Fatalf("Failed to get order at time zero, error: %s", err.Error())
	}
	if sorderAtZero.Size != 10 {
		t.Fatalf("Expected size at time zero to be 10, instead %d", sorderAtZero.Size)
	}
}

func TestMutatingTwoOrders(t *testing.T) {
	book := NewInMemoryOrderBook()
	orderOne := Order{ID: "foobar", Price: 100, Side: SIDE_BUY}
	orderTwo := Order{ID: "bazbar", Price: 100, Side: SIDE_SELL}
	book.PlaceOrder(orderOne, 10, time.Unix(0, 0))
	book.PlaceOrder(orderTwo, 10, time.Unix(0, 0))

	mut := &OrderStateChange{
		State: STATE_OPEN,
	}
	errs := book.MutateOrder("foobar", []OrderMutation{mut})
	if errs != nil {
		t.Fatalf("Unexpected error mutating order book: %s", errs)
	}
	sorder, _ := book.GetOrder("foobar")
	if sorder.State != STATE_OPEN {
		t.Fatalf("Mutation failed to apply. Expected state %s to be open", sorder.State)
	}
	sorder, _ = book.GetOrder("bazbar")
	if sorder.State != STATE_PENDING {
		t.Fatalf("Unexpected order modification. Expected state %s to be pending", sorder.State)
	}
}
