package coinbase

type CoinbaseMessage struct {
	CommandType string `json:"type"`
}

type SubscribeMessage struct {
	CoinbaseMessage
	ProductID string `json:"product_id"`
}

type OrderMessage struct {
	CoinbaseMessage
	Price    string `json:"price"`
	Sequence int    `json:"sequence"`
	Time     string `json:"time"`
	OrderID  string `json:"order_id"`
	Side     string `json:"side"`
}

type ReceivedOrderMessage struct {
	OrderMessage
	Size string `json:"size"`
}

type OpenOrderMessage struct {
	OrderMessage
	RemainingSize string `json:"remaining_size"`
}

type DoneOrderMessage struct {
	OrderMessage
	Reason        string `json:"reason"`
	RemainingSize string `json:"remaining_size"`
}

type ChangeOrderMessage struct {
	OrderMessage
	NewSize string `json:"new_size"`
	OldSize string `json:"old_size"`
}

type MatchOrderMessage struct {
	CoinbaseMessage
	Price        string `json:"price"`
	Sequence     int    `json:"sequence"`
	Side         string `json:"side"`
	MakerOrderID string `json:"maker_order_id"`
	TakerOrderID string `json:"taker_order_id"`
	TradeID      int    `json:"trade_id"`
}

type ErrorMessage struct {
	CoinbaseMessage
	Message string `json:"message"`
}

/*

	log.Print("Connecting to the Coinbase Exchange real-time websocket feed...")

	coinbase_url_raw := "wss://ws-feed.exchange.coinbase.com"
	coinbase_headers := http.Header{}
	coinbase_headers.Set("Origin", "http://www.jacobgreenleaf.com")

	ws_dialer := websocket.Dialer{}

	ws, _, err := ws_dialer.Dial(coinbase_url_raw, coinbase_headers)

	if err != nil {
		log.Fatalf("Error upgrading coinbase exchange feed connection to WebSocket: %s", err.Error())
	}

	log.Print("Subscribing to BTC/USD")

	subscribe_msg := &SubscribeMessage{CoinbaseMessage: CoinbaseMessage{CommandType: "subscribe"}, ProductID: "BTC-USD"}

	if err := ws.WriteJSON(subscribe_msg); err != nil {
		log.Fatalf("Error writing subscribe message: %s", err.Error())
	}

	log.Printf("Subscribed. Buffering...")

	order_buffer := list.New()

	for {
		var reader io.Reader

		_, reader, err := ws.NextReader()

		if err != nil {
			log.Printf("Error getting reader from websocket: %s", err.Error())
			break
		}

		decoder := json.NewDecoder(reader)
		for {
			var raw_order interface{}
			decoder.Decode(&raw_order)

			if raw_order == nil {
				break
			}

			bts, _ := json.Marshal(raw_order)

			var coinbase_msg CoinbaseMessage
			json.Unmarshal(bts, &coinbase_msg)

			//fmt.Printf("Received: %s.\n", coinbase_msg.CommandType)
			go statsdclient.Inc(coinbase_msg.CommandType, 1, 1)

			var order interface{}

			if "open" == coinbase_msg.CommandType {
				order = &OpenOrderMessage{}
			} else if "received" == coinbase_msg.CommandType {
				order = &ReceivedOrderMessage{}
			} else if "match" == coinbase_msg.CommandType {
				order = &MatchOrderMessage{}
			} else if "change" == coinbase_msg.CommandType {
				order = &ChangeOrderMessage{}
			} else if "done" == coinbase_msg.CommandType {
				order = &DoneOrderMessage{}
			} else if "error" == coinbase_msg.CommandType {
				var err ErrorMessage
				json.Unmarshal(bts, &err)
				fmt.Printf("Coinbase error: %s\n", err.Message)
				break
			} else {
				fmt.Printf("Unknown order type %s\n", coinbase_msg.CommandType)
				break
			}

			json.Unmarshal(bts, &order)

			order_buffer.PushFront(order)

			fmt.Printf("\r%d pending orders", order_buffer.Len())

			//			base_order := raw_order.(map[string]interface{})

			/*
				mutations := make([]*Hbase.Mutation, 0)

				mutations = append(mutations, goh.NewMutation("d:timestamp", []byte(base_order["time"].(string))))
				mutations = append(mutations, goh.NewMutation("d:side", []byte(base_order["side"].(string))))

				buf := new(bytes.Buffer)
				binary.Write(buf, binary.LittleEndian, base_order["sequence"].(float64))
				mutations = append(mutations, goh.NewMutation("d:sequence", buf.Bytes()))

				mutations = append(mutations, goh.NewMutation("d:price", []byte(base_order["price"].(string))))

				if coinbase_msg.CommandType != "change" {
					mutations = append(mutations, goh.NewMutation("d:status", []byte(coinbase_msg.CommandType)))
				}

				// TODO: THIS IS BROKEN
				switch ordr := order.(type) {
				case OpenOrderMessage:
					mutations = append(mutations, goh.NewMutation("d:size", []byte(ordr.RemainingSize)))
				case ReceivedOrderMessage:
					mutations = append(mutations, goh.NewMutation("d:size", []byte(ordr.Size)))
				case MatchOrderMessage:
					buf := new(bytes.Buffer)
					binary.Write(buf, binary.LittleEndian, ordr.TradeID)
					mutations = append(mutations, goh.NewMutation("d:trade_id", buf.Bytes()))
				case DoneOrderMessage:
					mutations = append(mutations, goh.NewMutation("d:size", []byte(ordr.RemainingSize)))
				case ChangeOrderMessage:
					mutations = append(mutations, goh.NewMutation("d:size", []byte(ordr.NewSize)))
				}

				if "match" != coinbase_msg.CommandType {
					go hbaseclient.MutateRow("coinbase_orders", []byte(base_order["order_id"].(string)), mutations, nil)
				} else {
					// Matches are two orders in one
					go hbaseclient.MutateRow("coinbase_orders", []byte(base_order["maker_order_id"].(string)), mutations, nil)
					go hbaseclient.MutateRow("coinbase_orders", []byte(base_order["taker_order_id"].(string)), mutations, nil)
				}
		}
	}
*/