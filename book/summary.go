package book

import "time"

// Calculate the sum of (order size * price) of all open orders
func CalculateTotalCentsInPlayInMemory(book *InMemoryOrderBook, t time.Time) int64 {
	var centsInPlay int64 = 0

	levels := book.GetPriceLevels()

	for _, level := range levels {
		orders := book.GetPriceLevelVersion(level, t)
		for _, order := range orders {
			if order.State == STATE_OPEN {
				centsInPlay += order.Price * order.Size
			}
		}
	}

	return centsInPlay
}

func CalculateNumberOfOpenOrdersInMemory(book *InMemoryOrderBook, t time.Time) int64 {
	var openOrders int64 = 0

	for orderId, _ := range book.Book {
		order, _ := book.GetOrderVersion(orderId, t)
		if order.State == STATE_OPEN {
			openOrders += 1
		}
	}

	return openOrders
}

func CalculateBidMedianAskSpreadInMemory(book *InMemoryOrderBook, t time.Time) (bid, median, ask, spread int64) {
	bid = -1
	ask = -1

	for orderId, _ := range book.Book {
		order, _ := book.GetOrderVersion(orderId, t)
		if order.State == STATE_OPEN {
			if order.Side == SIDE_BUY {
				if order.Price > bid {
					bid = order.Price
				}
			} else if order.Side == SIDE_SELL {
				if order.Price < ask || ask == -1 {
					ask = order.Price
				}
			}
		}
	}

	median = bid + ((ask - bid) / 2)
	spread = (ask - bid)

	return bid, median, ask, spread
}
