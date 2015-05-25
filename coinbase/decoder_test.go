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
				t.Fatalf("Expected open state mutation to be at %s instead %s", dt, mut.Time)
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
