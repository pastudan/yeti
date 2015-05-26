package coinbase

import "time"
import "testing"
import "bitbucket.org/jacobgreenleaf/yeti/book"

func TestDecodingReceiveOrders(t *testing.T) {
	cmds := Decode([]byte(`
		{
			"type": "received",
			"time": "2014-11-07T08:19:27.028459Z",
			"product_id": "BTC-USD",
			"sequence": 10,
			"order_id": "d50ec984-77a8-460a-b958-66f114b0de9b",
			"size": "0.10",
			"price": "0.10",
			"side": "buy"
		}
	`))

	if cmds == nil || len(cmds) != 1 {
		t.Fatal("Expected an order book command, but got nil")
	}

	cmd := cmds[0].(*book.OrderBookPlacementCommand)

	if cmd.Order.ID != "d50ec984-77a8-460a-b958-66f114b0de9b" {
		t.Fatalf("Expected order ID to be d50ec984-77a8-460a-b958-66f114b0de9b, instead %s", cmd.Order.ID)
	}
	if cmd.Order.Price != 10 {
		t.Fatalf("Expected price to be ten cents, instead %d", cmd.Order.Price)
	}
	if cmd.Order.Side != book.SIDE_BUY {
		t.Fatalf("Expected side to be %s, instead %s", book.SIDE_BUY, cmd.Order.Side)
	}
	if cmd.Size != int64(SATOSHI/10) {
		t.Fatalf("Expected size to be 10000000 satoshis, instead %d", cmd.Size)
	}
	dt := time.Date(2014, 11, 7, 8, 19, 27, 28459000, time.UTC)
	if !cmd.Time.Equal(dt) {
		t.Fatalf("Expected date to be %s", dt)
	}
}

func TestDecodingOpenOrders(t *testing.T) {
	cmds := Decode([]byte(`
		{
			"type": "open",
			"time": "2014-11-07T08:19:27.028459Z",
			"product_id": "BTC-USD",
			"sequence": 10,
			"order_id": "d50ec984-77a8-460a-b958-66f114b0de9b",
			"price": "200.2",
			"remaining_size": "1.00",
			"side": "sell"
		}
	`))

	if cmds == nil || len(cmds) != 1 {
		t.Fatal("Expected an order book command, but got nil")
	}

	cmd := cmds[0].(*book.OrderBookMutationCommand)

	if cmd.ID != "d50ec984-77a8-460a-b958-66f114b0de9b" {
		t.Fatalf("Expected order ID to be d50ec984-77a8-460a-b958-66f114b0de9b, instead %s", cmd.ID)
	}

	mutations := cmd.Mutations

	if mutations == nil || len(mutations) == 0 {
		t.Fatalf("Expected non-zero amount of mutations")
	}

	var hasStateMutation, hasSizeMutation bool = false, false
	dt := time.Date(2014, 11, 7, 8, 19, 27, 28459000, time.UTC)

	for _, mutation := range mutations {
		switch mut := mutation.(type) {

		case *book.OrderStateMutation:
			hasStateMutation = true
			if mut.State != book.STATE_OPEN {
				t.Fatalf("Expected open state mutation, instead %s", mut.State)
			}
			if !mut.Time.Equal(dt) {
				t.Fatalf("Expected open state mutation to be at %s instead %s", dt, mut.Time)
			}
		case *book.OrderSizeMutation:
			hasSizeMutation = true
			if mut.NewSize != int64(1*SATOSHI) {
				t.Fatalf("Expected size to be 10000000 satoshis, instead %d", mut.NewSize)
			}
			if !mut.Time.Equal(dt) {
				t.Fatalf("Expected size mutation to be at %s instead %s", dt, mut.Time)
			}
			break

		}
	}

	if !hasStateMutation {
		t.Fatalf("Expecting state mutation, was not in list %s", mutations)
	}

	if !hasSizeMutation {
		t.Fatalf("Expecting size mutation, was not in list %s", mutations)
	}

}

func TestDecodingDoneFilledOrders(t *testing.T) {
	cmds := Decode([]byte(`
		{
			"type": "done",
			"time": "2014-11-07T08:19:27.028459Z",
			"product_id": "BTC-USD",
			"sequence": 10,
			"price": "200.2",
			"order_id": "d50ec984-77a8-460a-b958-66f114b0de9b",
			"reason": "filled",
			"side": "sell",
			"remaining_size": "0"
		}
	`))

	if cmds == nil || len(cmds) != 1 {
		t.Fatal("Expected an order book command, but got nil")
	}

	cmd := cmds[0].(*book.OrderBookMutationCommand)

	if cmd.ID != "d50ec984-77a8-460a-b958-66f114b0de9b" {
		t.Fatalf("Expected order ID to be d50ec984-77a8-460a-b958-66f114b0de9b, instead %s", cmd.ID)
	}

	mutations := cmd.Mutations

	if mutations == nil || len(mutations) == 0 {
		t.Fatalf("Expected non-zero amount of mutations")
	}

	var hasStateMutation, hasSizeMutation bool = false, false
	dt := time.Date(2014, 11, 7, 8, 19, 27, 28459000, time.UTC)

	for _, mutation := range mutations {
		switch mut := mutation.(type) {

		case *book.OrderStateMutation:
			hasStateMutation = true
			if mut.State != book.STATE_FILLED {
				t.Fatalf("Expected filled state mutation, instead %s", mut.State)
			}
			if !mut.Time.Equal(dt) {
				t.Fatalf("Expected filled state mutation to be at %s instead %s", dt, mut.Time)
			}
		case *book.OrderSizeMutation:
			hasSizeMutation = true
			if mut.NewSize != 0 {
				t.Fatalf("Expected size to be 0 satoshis, instead %d", mut.NewSize)
			}
			if !mut.Time.Equal(dt) {
				t.Fatalf("Expected size mutation to be at %s instead %s", dt, mut.Time)
			}
			break

		}
	}

	if !hasStateMutation {
		t.Fatalf("Expecting state mutation, was not in list %s", mutations)
	}

	if !hasSizeMutation {
		t.Fatalf("Expecting size mutation, was not in list %s", mutations)
	}
}

func TestDecodingDoneCancelledOrders(t *testing.T) {
	cmds := Decode([]byte(`
		{
			"type": "done",
			"time": "2014-11-07T08:19:27.028459Z",
			"product_id": "BTC-USD",
			"sequence": 10,
			"price": "200.2",
			"order_id": "d50ec984-77a8-460a-b958-66f114b0de9b",
			"reason": "cancelled",
			"side": "sell",
			"remaining_size": "0.2"
		}
	`))

	if cmds == nil || len(cmds) != 1 {
		t.Fatal("Expected an order book command, but got nil")
	}

	cmd := cmds[0].(*book.OrderBookMutationCommand)

	if cmd.ID != "d50ec984-77a8-460a-b958-66f114b0de9b" {
		t.Fatalf("Expected order ID to be d50ec984-77a8-460a-b958-66f114b0de9b, instead %s", cmd.ID)
	}

	mutations := cmd.Mutations

	if mutations == nil || len(mutations) == 0 {
		t.Fatalf("Expected non-zero amount of mutations")
	}

	var hasStateMutation, hasSizeMutation bool = false, false
	dt := time.Date(2014, 11, 7, 8, 19, 27, 28459000, time.UTC)

	for _, mutation := range mutations {
		switch mut := mutation.(type) {

		case *book.OrderStateMutation:
			hasStateMutation = true
			if mut.State != book.STATE_VOID {
				t.Fatalf("Expected void state mutation, instead %s", mut.State)
			}
			if !mut.Time.Equal(dt) {
				t.Fatalf("Expected void state mutation to be at %s instead %s", dt, mut.Time)
			}
		case *book.OrderSizeMutation:
			hasSizeMutation = true
			if mut.NewSize != int64(SATOSHI/5) {
				t.Fatalf("Expected size to be 20000 satoshis, instead %d", mut.NewSize)
			}
			if !mut.Time.Equal(dt) {
				t.Fatalf("Expected size mutation to be at %s instead %s", dt, mut.Time)
			}
			break

		}
	}

	if !hasStateMutation {
		t.Fatalf("Expecting state mutation, was not in list %s", mutations)
	}

	if !hasSizeMutation {
		t.Fatalf("Expecting size mutation, was not in list %s", mutations)
	}
}

func TestDecodingMatchOrders(t *testing.T) {
	cmds := Decode([]byte(`
		{
			"type": "match",
			"trade_id": 10,
			"sequence": 50,
			"maker_order_id": "ac928c66-ca53-498f-9c13-a110027a60e8",
			"taker_order_id": "132fb6ae-456b-4654-b4e0-d681ac05cea1",
			"time": "2014-11-07T08:19:27.028459Z",
			"product_id": "BTC-USD",
			"size": "5.23512",
			"price": "400.23",
			"side": "sell"
		}
	`))

	if cmds == nil {
		t.Fatal("Expected two order book commands, but got nil")
	}

	if len(cmds) != 2 {
		t.Fatalf("Expected two order book commands, but got %d", len(cmds))
	}

	var cmd_maker, cmd_taker *book.OrderBookMutationCommand = nil, nil
	cmd_one := cmds[0].(*book.OrderBookMutationCommand)
	cmd_two := cmds[1].(*book.OrderBookMutationCommand)

	if cmd_one.ID == "ac928c66-ca53-498f-9c13-a110027a60e8" {
		cmd_maker = cmd_one
	} else if cmd_one.ID == "132fb6ae-456b-4654-b4e0-d681ac05cea1" {
		cmd_taker = cmd_one
	} else {
		t.Fatalf("Unexpected order id %s (%s)", cmd_one.ID, cmd_one)
	}

	if cmd_two.ID == "ac928c66-ca53-498f-9c13-a110027a60e8" {
		cmd_maker = cmd_two
	} else if cmd_two.ID == "132fb6ae-456b-4654-b4e0-d681ac05cea1" {
		cmd_taker = cmd_two
	} else {
		t.Fatalf("Unexpected order id %s (%s)", cmd_one.ID, cmd_one)
	}

	if cmd_maker == nil {
		t.Fatal("Expected maker order to be present")
	}
	if cmd_taker == nil {
		t.Fatal("Expected maker order to be present")
	}

	dt := time.Date(2014, 11, 7, 8, 19, 27, 28459000, time.UTC)

	maker_mutations := cmd_maker.Mutations
	taker_mutations := cmd_taker.Mutations

	if maker_mutations == nil || len(maker_mutations) != 1 || taker_mutations == nil || len(taker_mutations) != 1 {
		t.Fatalf("Expected one mutation")
	}

	maker_mutation := maker_mutations[0].(*book.OrderMatchMutation)

	if maker_mutation.Size != 523512 {
		t.Fatalf("Expected size to be 523512 satoshis, instead %d", maker_mutation.Size)
	}
	if !maker_mutation.Time.Equal(dt) {
		t.Fatalf("Expected size mutation to be at %s instead %s", dt, maker_mutation.Time)
	}
	if !maker_mutation.WasMaker {
		t.Fatal("Epected maker mutation to be a maker")
	}

	taker_mutation := taker_mutations[0].(*book.OrderMatchMutation)

	if taker_mutation.Size != 523512 {
		t.Fatalf("Expected size to be 523512 satoshis, instead %d", taker_mutation.Size)
	}
	if !taker_mutation.Time.Equal(dt) {
		t.Fatalf("Expected size mutation to be at %s instead %s", dt, taker_mutation.Time)
	}
	if taker_mutation.WasMaker {
		t.Fatal("Epected taker mutation to be a taker")
	}
	if taker_mutation.MakerID != "ac928c66-ca53-498f-9c13-a110027a60e8" {
		t.Fatalf("Expected taker maker id to be ac928c66-ca53-498f-9c13-a110027a60e8, instead %s", taker_mutation.MakerID)
	}
}

func TestDecodingChangeOrders(t *testing.T) {
	cmds := Decode([]byte(`
		{
			"type": "change",
			"time": "2014-11-07T08:19:27.028459Z",
			"sequence": 80,
			"order_id": "ac928c66-ca53-498f-9c13-a110027a60e8",
			"product_id": "BTC-USD",
			"new_size": "5.23512",
			"old_size": "12.234412",
			"price": "400.23",
			"side": "sell"
		}
	`))

	if cmds == nil {
		t.Fatal("Expected order book command, but got nil")
	}

	if len(cmds) != 1 {
		t.Fatalf("Expected one order book command, but got %d", len(cmds))
	}

	cmd := cmds[0].(*book.OrderBookMutationCommand)

	if cmd.ID != "ac928c66-ca53-498f-9c13-a110027a60e8" {
		t.Fatalf("Expected order id to be ac928c66-ca53-498f-9c13-a110027a60e8, instead %s", cmd.ID)
	}

	dt := time.Date(2014, 11, 7, 8, 19, 27, 28459000, time.UTC)

	mutations := cmd.Mutations

	if mutations == nil || len(mutations) != 1 {
		t.Fatalf("Expected one mutation")
	}

	mutation := mutations[0].(*book.OrderSizeMutation)

	if mutation.NewSize != 523512 {
		t.Fatalf("Expected size to be 523512 satoshis, instead %d", mutation.NewSize)
	}
	if !mutation.Time.Equal(dt) {
		t.Fatalf("Expected size mutation to be at %s instead %s", dt, mutation.Time)
	}
}

func TestDecodingError(t *testing.T) {
	cmds := Decode([]byte(`
		{
			"type": "error",
			"message": "error message"
		}
	`))

	if cmds != nil {
		t.Fatal("Expected error messages to return no commands")
	}
}