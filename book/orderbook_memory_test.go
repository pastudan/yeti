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

func TestPriceLevels(t *testing.T) {
	book := NewInMemoryOrderBook()

	order := Order{ID: "aaa", Price: 100, Side: SIDE_BUY}
	book.PlaceOrder(order, 10, time.Unix(0, 0))

	order = Order{ID: "bbb", Price: 200, Side: SIDE_BUY}
	book.PlaceOrder(order, 10, time.Unix(0, 0))

	order = Order{ID: "ccc", Price: 100, Side: SIDE_SELL}
	book.PlaceOrder(order, 10, time.Unix(0, 0))

	order = Order{ID: "ddd", Price: 100, Side: SIDE_BUY}
	book.PlaceOrder(order, 10, time.Unix(0, 0))

	prices := book.GetPriceLevel(100)
	if prices == nil {
		t.Fatal("Unexpected nil slice")
	}
	if len(prices) != 3 {
		t.Fatalf("Expected number of orders at price level 100 to be 3, instead %d", len(prices))
	}

	prices = book.GetPriceLevel(200)
	if prices == nil {
		t.Fatal("Unexpected nil slice")
	}
	if len(prices) != 1 {
		t.Fatalf("Expected number of orders at price level 200 to be 1, instead %d", len(prices))
	}

	prices = book.GetPriceLevel(10000000)
	if prices != nil && len(prices) != 0 {
		t.Fatal("Expected nil slice or empty slice")
	}

	book.MutateOrder("aaa", []OrderMutation{&OrderStateMutation{
		Time:  time.Unix(1, 0),
		State: STATE_VOID,
	}})

	prices = book.GetPriceLevel(100)
	if prices == nil {
		t.Fatal("Unexpected nil slice")
	}
	if len(prices) != 2 {
		t.Fatalf("Expected number of orders at price level 100 to be two after removing one, instead %d", len(prices))
	}
}

func TestMutatingSingleOrder(t *testing.T) {
	book := NewInMemoryOrderBook()
	order := Order{ID: "foobar", Price: 100, Side: SIDE_BUY}
	book.PlaceOrder(order, 10, time.Unix(0, 0))

	mut := &OrderStateMutation{
		State: STATE_OPEN,
		Time:  time.Unix(0, 0),
	}
	errs := book.MutateOrder("foobar", []OrderMutation{mut})
	if errs != nil {
		t.Fatalf("Unexpected error mutating order book: %s", errs)
	}
	sorder, err := book.GetOrder("foobar")
	if sorder.State != STATE_OPEN {
		t.Fatalf("Mutation failed to apply. Expected state %s to be %s", sorder.State, STATE_OPEN)
	}

	mut = &OrderStateMutation{
		State: STATE_OPEN,
		Time:  time.Unix(0, 0),
	}
	errs = book.MutateOrder("bazbar", []OrderMutation{mut})
	if errs == nil {
		t.Fatal("Expected state mutation on non-existent order to be invalid")
	}

	sizemut := &OrderSizeMutation{
		NewSize: 11,
		Time:    time.Unix(0, 0),
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

	sizemut_new := &OrderSizeMutation{
		NewSize: 20,
		Time:    time.Unix(1, 0),
	}
	sizemut_old := &OrderSizeMutation{
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

	match := &OrderMatchMutation{
		TradeID:  0,
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

	match = &OrderMatchMutation{
		TradeID:  1,
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

	match = &OrderMatchMutation{
		TradeID:  2,
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

	mut := &OrderStateMutation{
		State: STATE_VOID,
		Time:  time.Unix(1, 0),
	}
	book.MutateOrder("foobar", []OrderMutation{mut})

	sorder, err := book.GetOrder("foobar")
	if err != nil {
		t.Fatalf("Unexpected error when getting voided order: %s", err.Error())
	}
	if sorder.State != STATE_VOID {
		t.Fatalf("Unexpected state %s, expected voided", sorder.State)
	}
}

func TestOrderVersions(t *testing.T) {
	book := NewInMemoryOrderBook()
	order := Order{ID: "foobar", Price: 100, Side: SIDE_BUY}
	book.PlaceOrder(order, 10, time.Unix(0, 0))

	sorderAtZero, err := book.GetOrderVersion("foobar", time.Unix(0, 0))
	if err != nil {
		t.Fatalf("Failed to get order at time zero, error: %s", err.Error())
	}
	if sorderAtZero.Size != 10 {
		t.Fatalf("Expected size at time zero to be 10, instead %d", sorderAtZero.Size)
	}

	mut := &OrderSizeMutation{
		NewSize: 9,
		Time:    time.Unix(1, 0),
	}
	err = book.MutateOrder("foobar", []OrderMutation{mut})

	mut = &OrderSizeMutation{
		NewSize: 5,
		Time:    time.Unix(2, 0),
	}
	err = book.MutateOrder("foobar", []OrderMutation{mut})

	sorderAtZero, err = book.GetOrderVersion("foobar", time.Unix(0, 0))
	if err != nil {
		t.Fatalf("Failed to get order at time zero, error: %s", err.Error())
	}
	if sorderAtZero.Size != 10 {
		t.Fatalf("Expected size at time zero to be 10, instead %d", sorderAtZero.Size)
	}

	sorderAtOne, err := book.GetOrderVersion("foobar", time.Unix(1, 0))
	if err != nil {
		t.Fatalf("Failed to get order at time one, error: %s", err.Error())
	}
	if sorderAtOne.Size != 9 {
		t.Fatalf("Expected size at time one to be 9, instead %d", sorderAtOne.Size)
	}

	sorderAtTwo, err := book.GetOrderVersion("foobar", time.Unix(2, 0))
	if err != nil {
		t.Fatalf("Failed to get order at time one, error: %s", err.Error())
	}
	if sorderAtTwo.Size != 5 {
		t.Fatalf("Expected size at time one to be 5, instead %d", sorderAtTwo.Size)
	}

	sorderAtMinusOne, err := book.GetOrderVersion("foobar", time.Unix(-1, 0))
	if err != nil {
		t.Fatalf("Failed to get order at time minus one, error: %s", err.Error())
	}
	if sorderAtMinusOne.Size != 10 {
		t.Fatalf("Expected size at time minus one to be 10, instead %d", sorderAtMinusOne)
	}
}

func TestMutatingTwoOrders(t *testing.T) {
	book := NewInMemoryOrderBook()
	orderOne := Order{ID: "foobar", Price: 100, Side: SIDE_BUY}
	orderTwo := Order{ID: "bazbar", Price: 100, Side: SIDE_SELL}
	book.PlaceOrder(orderOne, 10, time.Unix(0, 0))
	book.PlaceOrder(orderTwo, 10, time.Unix(0, 0))

	mut := &OrderStateMutation{
		State: STATE_OPEN,
		Time:  time.Unix(0, 0),
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

func TestVacuuming(t *testing.T) {
	book := NewInMemoryOrderBook()

	order := Order{ID: "foobar", Price: 100, Side: SIDE_BUY}
	book.PlaceOrder(order, 10, time.Unix(0, 0))
	order = Order{ID: "bazbar", Price: 100, Side: SIDE_BUY}
	book.PlaceOrder(order, 10, time.Unix(0, 0))
	mut := &OrderStateMutation{
		State: STATE_VOID,
		Time:  time.Unix(1, 0),
	}
	book.MutateOrder("foobar", []OrderMutation{mut})
	book.Vacuum()

	sorder, err := book.GetOrder("foobar")
	if sorder != nil || err == nil {
		t.Fatal("Expected voided order to be removed from the book after vacuuming.")
	}

	sorder, err = book.GetOrder("bazbar")
	if sorder == nil || err != nil {
		t.Fatal("Expected pending order not to be removed by vacuum.")
	}
}

func TestLatestMutationTime(t *testing.T) {
	book := NewInMemoryOrderBook()

	order := Order{ID: "foobar", Price: 100, Side: SIDE_BUY}
	book.PlaceOrder(order, 10, time.Unix(0, 0))

	if !book.LatestMutationTime.Equal(time.Unix(0, 0)) {
		t.Fatalf("Expected latest mutation time to be %s, instead %s", time.Unix(0, 0), book.LatestMutationTime)
	}

	mut := &OrderStateMutation{
		State: STATE_OPEN,
		Time:  time.Unix(1, 0),
	}
	book.MutateOrder("foobar", []OrderMutation{mut})

	if !book.LatestMutationTime.Equal(time.Unix(1, 0)) {
		t.Fatalf("Expected latest mutation time of the order book to be %s, instead %s", time.Unix(1, 0), book.LatestMutationTime)
	}

	sorder, _ := book.GetOrder("foobar")
	if !sorder.LatestMutationTime.Equal(time.Unix(1, 0)) {
		t.Fatalf("Expected latest mutation time of the order to be %s, instead %s", time.Unix(1, 0), sorder.LatestMutationTime)
	}

	sorder, _ = book.GetOrderVersion("foobar", time.Unix(0, 0))
	if !sorder.LatestMutationTime.Equal(time.Unix(0, 0)) {
		t.Fatalf("Expected latest mutation time of the order at t=0 to be %s, instead %s", time.Unix(0, 0), sorder.LatestMutationTime)
	}

	book.MutateOrder("foobar", []OrderMutation{&OrderSizeMutation{
		NewSize: 5,
		Time:    time.Unix(0, 0),
	}})

	if !book.LatestMutationTime.Equal(time.Unix(1, 0)) {
		t.Fatalf("Expected latest mutation time of the order book to be %s, instead %s", time.Unix(1, 0), book.LatestMutationTime)
	}
}

func TestOrderBookCommands(t *testing.T) {
	var err error

	book := NewInMemoryOrderBook()

	placementCmd := OrderBookPlacementCommand{
		Order: Order{
			ID:    "foobar",
			Price: 100,
			Side:  SIDE_SELL,
		},
		Size: 10,
		Time: time.Unix(0, 0),
	}

	err = placementCmd.Apply(book)
	if err != nil {
		t.Fatalf("Unexpected error when executing order placement command: %s", err.Error())
	}

	order, err := book.GetOrder("foobar")
	if order == nil || err == errOrderDoesNotExist {
		t.Fatalf("Placement command failed to place an order.")
	} else if err != nil {
		t.Fatalf("Unexpected error when getting order: %s", err.Error())
	}

	mutationCmd := OrderBookMutationCommand{
		ID: "foobar",
		Mutations: []OrderMutation{&OrderStateMutation{
			State: STATE_OPEN,
			Time:  time.Unix(1, 0),
		}},
	}

	err = mutationCmd.Apply(book)
	if err != nil {
		t.Fatalf("Unexpected error when executing order mutation command: %s", err.Error())
	}

	order, err = book.GetOrder("foobar")
	if order == nil || err == errOrderDoesNotExist {
		t.Fatalf("Mutation command unexpectedly deleted order? Error not found.")
	} else if err != nil {
		t.Fatalf("Unexpected error when getting order: %s", err.Error())
	}
	if order.State != STATE_OPEN {
		t.Fatalf("Mutation command failed to change state. Expected %s to be %s", order.State, STATE_OPEN)
	}
}
