package coinbase

import "testing"

func TestDecodingRESTOrderBook(t *testing.T) {
	response := []byte(`
		{
			"sequence": 3,
			"bids": [
				[ "1.00", "0.01", "aaaa" ],
				[ "1.01", "0.01", "bbbb" ],
				[ "1.02", "0.01", "cccc" ]
			],
			"asks": [
				[ "1.10", "0.01", "dddd" ],
				[ "1.11", "0.01", "cccc" ]
			]
		}
	`)

	seq, batch, err := DecodeRESTOrderBook(response)

	if err != nil {
		t.Fatalf("Unexpected error decoding order book: %s", err.Error())
	}

	if batch == nil {
		t.Fatal("Expected non-nil batch of commands")
	}

	if seq != int64(3) {
		t.Fatalf("Expected sequence number to be 3, instead %d", seq)
	}

	if len(batch.Commands) != 5 {
		t.Fatalf("Expected number of commands to be 5, instead %d", len(batch.Commands))
	}

	if batch.Sequence != int(3) {
		t.Fatalf("Expected sequence number to be 3, instead %d", batch.Sequence)
	}
}
